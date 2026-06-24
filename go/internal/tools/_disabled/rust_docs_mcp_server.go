package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchRustDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	maxResults, _ :=getInt(args, "max_results")
	if maxResults == 0 {
		maxResults = 5
	}
	u := fmt.Sprintf("https://doc.rust-lang.org/search?q=%s&limit=%d", url.QueryEscape(query), maxResults)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch docs: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return success(string(body))
}

	return success(fmt.Sprintf("Search results: %v", result))
}