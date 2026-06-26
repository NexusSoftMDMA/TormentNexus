package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleOrchestrate_mcp_orchestrator(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter is required")
}

	method, _ :=getString(args, "method")
	if method == "" {
		method = "GET"
	}
	req, e := http.NewRequestWithContext(ctx, method, url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %s", e.Error()))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %s", e.Error()))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %s", e.Error()))
}

	return ok(fmt.Sprintf("Response status %d: %s", resp.StatusCode, string(body)))
}