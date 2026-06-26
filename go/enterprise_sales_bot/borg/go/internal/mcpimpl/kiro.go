package mcpimpl

import (
	"context"
	"fmt"
	"net/http"
)

func HandleX_kiro(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url argument is required")
}

	req, e := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return ok(fmt.Sprintf("unreachable: %v", e))
}

	resp.Body.Close()
	return ok(fmt.Sprintf("reachable (status %d)", resp.StatusCode))
}