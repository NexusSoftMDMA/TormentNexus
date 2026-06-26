package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type searchResult_giskard_search struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

func HandleSearch_giskard_search(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := "https://api.giskard.ai/v1/search?q=" + query
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

	var result []searchResult
	if e = json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok("search completed")
}