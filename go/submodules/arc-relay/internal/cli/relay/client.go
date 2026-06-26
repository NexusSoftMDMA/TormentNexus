package relay

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Server represents an MCP server from the relay API.
type Server struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
	ServerType    string `json:"server_type"`
	Status        string `json:"status"`
	Health        string `json:"health,omitempty"`          // "healthy", "unhealthy", "unknown", or ""
	HealthCheckAt string `json:"health_check_at,omitempty"` // ISO 8601 timestamp
	HealthError   string `json:"health_error,omitempty"`    // error message when unhealthy
}

// Client is an API client for an Arc Relay instance.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new relay API client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ListServers fetches all servers from the relay API.
func (c *Client) ListServers() ([]Server, error) {
	url := c.BaseURL + "/api/servers"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connecting to relay at %s: %w", c.BaseURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// continue
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("authentication failed (401) — check your API key")
	case http.StatusForbidden:
		return nil, fmt.Errorf("access denied (403) — your API key may lack permissions")
	default:
		return nil, fmt.Errorf("relay returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var servers []Server
	if err := json.Unmarshal(body, &servers); err != nil {
		return nil, fmt.Errorf("parsing server list: %w", err)
	}

	return servers, nil
}

// ListRunningServers fetches servers and filters to only those with status "running".
func (c *Client) ListRunningServers() ([]Server, error) {
	all, err := c.ListServers()
	if err != nil {
		return nil, err
	}

	var running []Server
	for _, s := range all {
		if s.Status == "running" {
			running = append(running, s)
		}
	}
	return running, nil
}

// ServerProxyURL returns the MCP proxy URL for a given server name.
func (c *Client) ServerProxyURL(serverName string) string {
	return c.BaseURL + "/mcp/" + serverName
}

// IsRelayURL checks if a URL belongs to this relay instance's proxy.
func (c *Client) IsRelayURL(url string) bool {
	return strings.HasPrefix(url, c.BaseURL+"/mcp/")
}

// ServerNameFromURL extracts the server name from a relay proxy URL.
// Returns empty string if the URL is not a relay proxy URL.
func (c *Client) ServerNameFromURL(url string) string {
	prefix := c.BaseURL + "/mcp/"
	if !strings.HasPrefix(url, prefix) {
		return ""
	}
	name := strings.TrimPrefix(url, prefix)
	// Strip any trailing path segments
	if idx := strings.Index(name, "/"); idx >= 0 {
		name = name[:idx]
	}
	return name
}
