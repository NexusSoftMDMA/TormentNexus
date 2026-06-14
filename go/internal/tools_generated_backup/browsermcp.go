package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleNavigate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
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

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return ok(fmt.Sprintf("navigated to %s (status %d)", url, resp.StatusCode))
}