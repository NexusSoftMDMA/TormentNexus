package mcpimpl

import (
	"context"
	"io"
	"net/http"
	"time"
)

func HandleFetchURL_oraios(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	return ok(string(body))
}

func HandleGetTime_oraios(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now()
	return ok(now.Format(time.RFC3339))
}