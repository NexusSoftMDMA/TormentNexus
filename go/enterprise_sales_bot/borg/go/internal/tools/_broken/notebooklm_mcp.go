package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

func HandleNotebooklmSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	maxResults, _ :=getInt(args, "maxResults")
	base := "https://notebooklm.google.com/api/search"
	params := url.Values{}
	params.Set("q", query)
	if maxResults > 0 {
		params.Set("max_results", strconv.Itoa(maxResults))

	req, e := http.NewRequestWithContext(ctx, "GET", base+"?"+params.Encode(), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok("Search results retrieved")
}
}