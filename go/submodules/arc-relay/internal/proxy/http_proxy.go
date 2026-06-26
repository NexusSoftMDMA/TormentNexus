package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/comma-compliance/arc-relay/internal/mcp"
)

// HTTPProxy forwards MCP requests to an HTTP-based MCP server.
type HTTPProxy struct {
	targetURL  string
	mu         sync.Mutex
	sessionID  string
	httpClient *http.Client
}

// NewHTTPProxy creates a proxy to an HTTP MCP server.
func NewHTTPProxy(targetURL string) *HTTPProxy {
	return &HTTPProxy{
		targetURL:  targetURL,
		httpClient: &http.Client{},
	}
}

// SendNotification sends a fire-and-forget notification to the HTTP backend.
func (p *HTTPProxy) SendNotification(n *mcp.Notification) error {
	body, _ := json.Marshal(n)
	httpReq, err := http.NewRequest("POST", p.targetURL, bytes.NewReader(body))
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
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}

// Send forwards an MCP request to the HTTP backend.
func (p *HTTPProxy) Send(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	p.mu.Lock()
	sendSID := p.sessionID
	p.mu.Unlock()
	if sendSID != "" {
		httpReq.Header.Set("Mcp-Session-Id", sendSID)
	}

	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending request to %s: %w", p.targetURL, err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("backend returned status %d", httpResp.StatusCode)
	}

	// Capture session ID if provided
	if newSID := httpResp.Header.Get("Mcp-Session-Id"); newSID != "" {
		p.mu.Lock()
		p.sessionID = newSID
		p.mu.Unlock()
	}

	return parseHTTPResponse(httpResp)
}
