package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

func HandleWebSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	searchURL := "https://api.duckduckgo.com/?q=" + url.QueryEscape(query) + "&format=json&no_html=1&skip_disambig=1"
	resp, e := http.DefaultClient.Get(searchURL)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		AbstractText string `json:"AbstractText"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	if result.AbstractText != "" {
		return success(result.AbstractText)
}

	return ok("No results found")
}