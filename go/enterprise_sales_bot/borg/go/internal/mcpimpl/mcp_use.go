package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleUse(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	return ok(fmt.Sprintf("Status: %d\nBody: %s", resp.StatusCode, string(body)))
}