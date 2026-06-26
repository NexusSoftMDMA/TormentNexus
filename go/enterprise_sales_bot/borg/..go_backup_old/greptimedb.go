package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleGreptimedbQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "http://localhost:4000/v1/sql"
	}
	if query == "" {
		return err("query is required")
}

	data := url.Values{}
	data.Set("sql", query)
	resp, e := http.DefaultClient.PostForm(baseURL, data)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}