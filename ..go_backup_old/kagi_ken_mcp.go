package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleKagiSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiKey := os.Getenv("KAGI_API_KEY")
	if apiKey == "" {
		return err("KAGI_API_KEY not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://kagi.com/api/v0/search?q="+url.QueryEscape(query), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
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