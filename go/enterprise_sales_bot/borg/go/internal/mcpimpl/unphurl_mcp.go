package mcpimpl

import (
	"context"
	"net/http"
)

func HandleUnphurl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("URL is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("Invalid URL: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	finalURL := resp.Request.URL.String()
	return ok("Final URL: " + finalURL)
}