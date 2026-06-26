package mcpimpl

import (
	"context"
	"fmt"
	"net/http"
)

func HandlePing_imcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return ok("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return success(fmt.Sprintf("got status %d", resp.StatusCode))
}

	return err(fmt.Sprintf("unexpected status %d", resp.StatusCode))
}