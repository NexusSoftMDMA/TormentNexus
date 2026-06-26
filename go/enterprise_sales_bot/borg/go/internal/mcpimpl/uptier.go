package mcpimpl

import (
	"context"
	"fmt"
	"net/http"
)

func HandleCheckStatus_uptier(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
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
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("server returned %d", resp.StatusCode))
}

	return success("Server is up")
}