package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// HandleQueryRegistry queries the MCP Registry SDK API.
func HandleQueryRegistry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON response: " + e.Error())
	}
	return success(data)
}