package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"
)

// OAuthDiscovery holds the results of .well-known OAuth discovery.
type OAuthDiscovery struct {
	AuthURL               string   `json:"auth_url"`
	TokenURL              string   `json:"token_url"`
	RegistrationEndpoint  string   `json:"registration_endpoint,omitempty"`
	ScopesSupported       []string `json:"scopes_supported,omitempty"`
	ClientID              string   `json:"client_id,omitempty"`
	ClientSecret          string   `json:"client_secret,omitempty"`
	RegisteredRedirectURI string   `json:"registered_redirect_uri,omitempty"`
}

// ClientRegistration holds the result of dynamic client registration.
type ClientRegistration struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// ssrfSafeDialer wraps the default dialer to check resolved IPs at connection
// time, preventing DNS rebinding attacks (TOCTOU bypass of pre-request checks).
var ssrfSafeDialer = &net.Dialer{Timeout: 10 * time.Second}

func ssrfSafeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("dns resolution failed: %w", err)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses found for host %q", host)
	}
	for _, ip := range ips {
		if isPrivateIP(ip.IP) {
			return nil, fmt.Errorf("connection to private/loopback address blocked")
		}
	}
	// Dial the first resolved address to avoid re-resolving (all IPs verified safe above)
	return ssrfSafeDialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
}

func isPrivateIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

// discoveryClient is an HTTP client for OAuth discovery with defence-in-depth
// against SSRF: a custom DialContext rejects private IPs at connection time
// (closing the DNS rebinding TOCTOU window), and CheckRedirect limits hops.
var discoveryClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: func() *http.Transport {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.DialContext = ssrfSafeDialContext
		return t
	}(),
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("too many redirects")
		}
		return nil
	},
}

// DiscoverOAuth probes .well-known endpoints to auto-discover OAuth configuration.
// Returns nil, nil if no OAuth is found (not an error).
func DiscoverOAuth(ctx context.Context, serverURL string) (*OAuthDiscovery, error) {
	parsed, err := url.Parse(serverURL)
	if err != nil {
		return nil, nil
	}
	origin := parsed.Scheme + "://" + parsed.Host

	// Step 1: Probe /.well-known/oauth-protected-resource to find the auth server
	authServer, err := probeProtectedResource(ctx, origin)
	if err != nil || authServer == "" {
		// Fallback: try the origin itself as the authorization server.
		// Some providers (e.g. Shortcut) only expose oauth-authorization-server
		// without the oauth-protected-resource endpoint.
		slog.Debug("OAuth discovery: no protected-resource found, falling back to origin as auth server", "origin", origin)
		authServer = origin
	}

	// Step 2: Probe /.well-known/oauth-authorization-server on the auth server
	slog.Debug("OAuth discovery: probing authorization server", "auth_server", authServer)
	discovery, err := probeAuthorizationServer(ctx, authServer)
	if err != nil || discovery == nil {
		return nil, nil
	}

	return discovery, nil
}

// RegisterClient performs dynamic client registration at the given endpoint.
func RegisterClient(ctx context.Context, registrationEndpoint, callbackURL string) (*ClientRegistration, error) {
	body := map[string]any{
		"client_name":                "Arc Relay",
		"redirect_uris":              []string{callbackURL},
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "client_secret_post",
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling registration request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", registrationEndpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating registration request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := discoveryClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("registration request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading registration response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("registration endpoint returned %d: %s", resp.StatusCode, string(respBody))
	}

	var reg struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}
	if err := json.Unmarshal(respBody, &reg); err != nil {
		return nil, fmt.Errorf("parsing registration response: %w", err)
	}

	if reg.ClientID == "" {
		return nil, fmt.Errorf("no client_id in registration response")
	}

	return &ClientRegistration{
		ClientID:     reg.ClientID,
		ClientSecret: reg.ClientSecret,
	}, nil
}

// probeProtectedResource fetches /.well-known/oauth-protected-resource and returns the first authorization server.
func probeProtectedResource(ctx context.Context, origin string) (string, error) {
	u := origin + "/.well-known/oauth-protected-resource"
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := discoveryClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", nil // no OAuth, not an error
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var resource struct {
		AuthorizationServers []string `json:"authorization_servers"`
	}
	if err := json.Unmarshal(body, &resource); err != nil {
		return "", nil
	}

	if len(resource.AuthorizationServers) == 0 {
		return "", nil
	}
	return resource.AuthorizationServers[0], nil
}

// probeAuthorizationServer fetches /.well-known/oauth-authorization-server and returns OAuth endpoints.
func probeAuthorizationServer(ctx context.Context, authServer string) (*OAuthDiscovery, error) {
	parsed, err := url.Parse(authServer)
	if err != nil {
		return nil, nil
	}
	origin := parsed.Scheme + "://" + parsed.Host

	u := origin + "/.well-known/oauth-authorization-server"
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := discoveryClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var as struct {
		AuthorizationEndpoint string   `json:"authorization_endpoint"`
		TokenEndpoint         string   `json:"token_endpoint"`
		RegistrationEndpoint  string   `json:"registration_endpoint"`
		ScopesSupported       []string `json:"scopes_supported"`
	}
	if err := json.Unmarshal(body, &as); err != nil {
		return nil, nil
	}

	if as.AuthorizationEndpoint == "" || as.TokenEndpoint == "" {
		return nil, nil
	}

	return &OAuthDiscovery{
		AuthURL:              as.AuthorizationEndpoint,
		TokenURL:             as.TokenEndpoint,
		RegistrationEndpoint: as.RegistrationEndpoint,
		ScopesSupported:      as.ScopesSupported,
	}, nil
}
