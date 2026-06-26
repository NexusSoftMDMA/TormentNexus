package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/comma-compliance/arc-relay/internal/mcp"
	"github.com/comma-compliance/arc-relay/internal/oauth"
	"github.com/comma-compliance/arc-relay/internal/store"
)

// RemoteProxy forwards MCP requests to a remote MCP server with auth.
type RemoteProxy struct {
	serverID       string
	config         store.RemoteConfig
	mu             sync.Mutex
	sessionID      string
	lastOAuthToken string // tracks token to detect proactive refreshes
	httpClient     *http.Client
	oauthManager   *oauth.Manager
}

// NewRemoteProxy creates a proxy to a remote MCP server.
func NewRemoteProxy(serverID string, config store.RemoteConfig, oauthMgr *oauth.Manager) *RemoteProxy {
	return &RemoteProxy{
		serverID:     serverID,
		config:       config,
		httpClient:   &http.Client{},
		oauthManager: oauthMgr,
	}
}

// applyAuth sets the appropriate auth header on the request.
// Returns true if an OAuth token was proactively refreshed (i.e., the token
// changed since the last request), which signals that the MCP session may
// have been invalidated by the remote server.
func (p *RemoteProxy) applyAuth(ctx context.Context, req *http.Request) (tokenRefreshed bool, err error) {
	switch p.config.Auth.Type {
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+p.config.Auth.Token)
	case "api_key":
		name := p.config.Auth.HeaderName
		if name == "" {
			name = "X-API-Key"
		}
		req.Header.Set(name, p.config.Auth.Token)
	case "private_url", "none", "":
		// No header needed
	case "oauth":
		if p.oauthManager == nil {
			return false, fmt.Errorf("OAuth manager not configured")
		}
		token, err := p.oauthManager.GetAccessToken(ctx, p.serverID)
		if err != nil {
			return false, fmt.Errorf("getting OAuth token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		// Detect proactive token refresh by comparing with last-used token
		p.mu.Lock()
		if p.lastOAuthToken != "" && p.lastOAuthToken != token {
			tokenRefreshed = true
		}
		p.lastOAuthToken = token
		p.mu.Unlock()
	}
	return tokenRefreshed, nil
}

// SendNotification sends a fire-and-forget notification to the remote server.
func (p *RemoteProxy) SendNotification(n *mcp.Notification) error {
	body, _ := json.Marshal(n)
	httpReq, err := http.NewRequest("POST", p.config.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	p.mu.Lock()
	sid := p.sessionID
	p.mu.Unlock()
	if sid != "" {
		httpReq.Header.Set("Mcp-Session-Id", sid)
	}
	if _, err := p.applyAuth(context.Background(), httpReq); err != nil {
		return err
	}
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}

// Send forwards an MCP request to the remote server.
func (p *RemoteProxy) Send(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
	resp, statusCode, err := p.doSend(ctx, req)

	// If 401 and OAuth, refresh the token, re-establish the MCP session,
	// and retry. Some servers (e.g. Shortcut) invalidate the MCP session
	// when the OAuth token changes, so we must re-initialize after refresh.
	if statusCode == http.StatusUnauthorized && p.config.Auth.Type == "oauth" && p.oauthManager != nil {
		slog.Warn("OAuth 401, refreshing token and re-initializing session", "server_id", p.serverID)
		if refreshErr := p.oauthManager.ForceRefresh(ctx, p.serverID); refreshErr != nil {
			return nil, fmt.Errorf("token refresh after 401 failed: %w", refreshErr)
		}
		p.mu.Lock()
		p.sessionID = ""
		p.lastOAuthToken = "" // reset so retry doSend doesn't trigger redundant reinitialize
		p.mu.Unlock()

		// Re-initialize to establish a fresh session with the new token
		if initErr := p.reinitialize(ctx); initErr != nil {
			slog.Warn("session re-initialize failed after token refresh", "server_id", p.serverID, "err", initErr)
		}

		resp, _, err = p.doSend(ctx, req)
	}

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// reinitialize sends an MCP initialize request to establish a fresh session.
func (p *RemoteProxy) reinitialize(ctx context.Context) error {
	id, _ := json.Marshal(0)
	params, _ := json.Marshal(map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]string{"name": "arc-relay", "version": "0.1.0"},
	})
	req := &mcp.Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "initialize",
		Params:  params,
	}
	_, _, err := p.doSend(ctx, req)
	return err
}

func (p *RemoteProxy) doSend(ctx context.Context, req *mcp.Request) (*mcp.Response, int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, 0, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.URL, bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")

	// Apply auth first — may trigger proactive OAuth token refresh
	tokenRefreshed, err := p.applyAuth(ctx, httpReq)
	if err != nil {
		return nil, 0, fmt.Errorf("applying auth: %w", err)
	}

	// If OAuth token was proactively refreshed, re-establish the MCP session
	// before sending. Some servers (e.g. Shortcut) invalidate the MCP session
	// when the OAuth token changes, so sending with the old session ID fails.
	if tokenRefreshed {
		p.mu.Lock()
		hadSession := p.sessionID != ""
		p.sessionID = ""
		p.mu.Unlock()
		if hadSession {
			slog.Info("OAuth token proactively refreshed, re-establishing session", "server_id", p.serverID)
			if initErr := p.reinitialize(ctx); initErr != nil {
				slog.Warn("session re-initialize after proactive refresh failed", "server_id", p.serverID, "err", initErr)
			}
		}
	}

	// initialize always starts a fresh session — never send a stale session ID.
	// This ensures Enumerate, reinitialize, and any other caller that sends
	// initialize work correctly without needing to clear the session first.
	if req.Method == "initialize" {
		p.mu.Lock()
		p.sessionID = ""
		p.mu.Unlock()
	}

	// Set session ID (may have been updated by reinitialize above)
	p.mu.Lock()
	sid := p.sessionID
	p.mu.Unlock()
	if sid != "" {
		httpReq.Header.Set("Mcp-Session-Id", sid)
	}

	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, 0, fmt.Errorf("sending request to %s: %w", p.config.URL, err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode == http.StatusUnauthorized {
		return nil, http.StatusUnauthorized, fmt.Errorf("remote server returned status 401")
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, httpResp.StatusCode, fmt.Errorf("remote server returned status %d", httpResp.StatusCode)
	}

	// Capture session ID if provided
	if newSID := httpResp.Header.Get("Mcp-Session-Id"); newSID != "" {
		p.mu.Lock()
		p.sessionID = newSID
		p.mu.Unlock()
	}

	resp, err := parseHTTPResponse(httpResp)
	return resp, http.StatusOK, err
}
