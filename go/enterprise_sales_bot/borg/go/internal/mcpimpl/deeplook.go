package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleLookup_deeplook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiURL := "https://en.wikipedia.org/api/rest_v1/page/summary/" + url.QueryEscape(query)
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", e))
}

	defer resp.Body.Close()
	var result struct {
		Extract string `json:"extract"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("JSON decode failed: %v", e))
}

	if result.Extract == "" {
		return err("no extract found")
}

	return ok(result.Extract)
}