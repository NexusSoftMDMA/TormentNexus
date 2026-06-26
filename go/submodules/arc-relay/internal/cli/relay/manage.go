package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// parseErrorBody attempts to extract an "error" field from a JSON response body.
// Returns the error string if found, otherwise returns the raw body as a string.
func parseErrorBody(body []byte) string {
	var errResp struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return errResp.Error
	}
	if len(body) > 0 {
		return string(body)
	}
	return ""
}

// handleErrorResponse returns an appropriate error for non-success HTTP responses.
// It parses the response body for structured error messages from the relay.
func handleErrorResponse(resp *http.Response, body []byte, context string) error {
	detail := parseErrorBody(body)

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication failed (401) — check your API key")
	case http.StatusForbidden:
		msg := "access denied (403)"
		if detail != "" {
			msg += " — " + detail
		}
		return fmt.Errorf("%s", msg)
	case http.StatusNotFound:
		if detail != "" {
			return fmt.Errorf("not found: %s", detail)
		}
		return fmt.Errorf("%s not found", context)
	case http.StatusConflict:
		if detail != "" {
			return fmt.Errorf("conflict: %s", detail)
		}
		return fmt.Errorf("%s already exists", context)
	case http.StatusBadRequest:
		if detail != "" {
			return fmt.Errorf("bad request: %s", detail)
		}
		return fmt.Errorf("bad request")
	default:
		if detail != "" {
			return fmt.Errorf("relay returned HTTP %d: %s", resp.StatusCode, detail)
		}
		return fmt.Errorf("relay returned HTTP %d", resp.StatusCode)
	}
}

// CreateServer creates a new server on the relay instance.
// Requires the API key to have admin or write access.
// Corresponds to POST /api/servers on the relay.
func (c *Client) CreateServer(req *CreateServerRequest) (*ServerDetail, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := c.BaseURL + "/api/servers"
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("connecting to relay at %s: %w", c.BaseURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, handleErrorResponse(resp, respBody, fmt.Sprintf("server %q", req.Name))
	}

	var detail ServerDetail
	if err := json.Unmarshal(respBody, &detail); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &detail, nil
}

// DeleteServer deletes a server from the relay instance.
// Corresponds to DELETE /api/servers/{id} on the relay.
func (c *Client) DeleteServer(serverID string) error {
	url := c.BaseURL + "/api/servers/" + serverID
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("connecting to relay: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return handleErrorResponse(resp, body, fmt.Sprintf("server %q", serverID))
}

// StartServer starts a server on the relay instance.
// Corresponds to POST /api/servers/{id}/start on the relay.
func (c *Client) StartServer(serverID string) error {
	return c.serverAction(serverID, "start")
}

// StopServer stops a server on the relay instance.
// Corresponds to POST /api/servers/{id}/stop on the relay.
func (c *Client) StopServer(serverID string) error {
	return c.serverAction(serverID, "stop")
}

func (c *Client) serverAction(serverID, action string) error {
	url := fmt.Sprintf("%s/api/servers/%s/%s", c.BaseURL, serverID, action)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("connecting to relay: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return handleErrorResponse(resp, body, fmt.Sprintf("server %q", serverID))
}

// GetServer fetches a single server's details by ID.
// Corresponds to GET /api/servers/{id} on the relay.
func (c *Client) GetServer(serverID string) (*ServerDetail, error) {
	url := c.BaseURL + "/api/servers/" + serverID
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connecting to relay: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, handleErrorResponse(resp, body, fmt.Sprintf("server %q", serverID))
	}

	var detail ServerDetail
	if err := json.Unmarshal(body, &detail); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &detail, nil
}
