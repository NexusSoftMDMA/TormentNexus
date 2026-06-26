package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type searchResult_smartest_tv struct {
	Show struct {
		Name    string `json:"name"`
		Summary string `json:"summary"`
	} `json:"show"`
}

func HandleSearch_smartest_tv(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	url := fmt.Sprintf("https://api.tvmaze.com/search/shows?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch: %v", e))
}

	defer resp.Body.Close()
	var results []searchResult
	if e := json.NewDecoder(resp.Body).Decode(&results); e != nil {
		return err(fmt.Sprintf("failed to decode: %v", e))
}

	if len(results) == 0 {
		return ok(fmt.Sprintf("No shows found for '%s'", query))
}

	show := results[0].Show
	return ok(fmt.Sprintf("Show: %s\nSummary: %s", show.Name, show.Summary))
}