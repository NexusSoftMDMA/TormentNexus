package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearch_searchcraft_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	resp, e := http.DefaultClient.Get("https://api.searchcraft.io/search?q=" + url.QueryEscape(query))
	if e != nil {
		return err("search request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("search API returned status %d", resp.StatusCode))
}

	var results map[string]interface{}
	json.Unmarshal(body, &results)
	out, _ := json.Marshal(results)
	return ok(string(out))
}