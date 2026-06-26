package tools

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func HandleGsc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var body strings.Builder
	if _, e = body.ReadFrom(resp.Body); e != nil {
		return err(fmt.Sprintf("failed to read body: %v", e))
}

	return ok(fmt.Sprintf("Status: %s, Body: %s", resp.Status, body.String()))
}

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query is required")
}

	return ok("Searching for: " + q)
}