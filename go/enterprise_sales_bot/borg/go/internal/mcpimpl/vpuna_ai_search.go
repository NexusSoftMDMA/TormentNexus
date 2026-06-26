package mcpimpl

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleSearch_vpuna_ai_search(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	reqURL := "https://api.vpuna.ai/search?q=" + url.QueryEscape(query)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response body")
}

	return ok(string(body))
}