package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleStatus_neat_no_pulse_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read body: %v", e))
}

	return ok(fmt.Sprintf("Status %d: %s", resp.StatusCode, string(body)))
}

func HandleSendPulse(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	message, _ :=getString(args, "message")
	if url == "" {
		return err("url is required")
}

	payload := map[string]string{"message": message}
	data, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Body = io.NopCloser(bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	return success(fmt.Sprintf("Pulse sent, status %d", resp.StatusCode))
}