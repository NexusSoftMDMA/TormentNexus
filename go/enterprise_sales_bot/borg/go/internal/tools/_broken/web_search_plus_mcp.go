package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func HandleWebSearchPlus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	maxResults, _ :=getInt(args, "maxResults")
	if maxResults <= 0 || maxResults > 10 {
		maxResults = 5
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("no_html", "1")
	params.Set("skip_disambig", "1")

	urlStr := fmt.Sprintf("https://api.duckduckgo.com/?%s", params.Encode())
	resp, e := http.DefaultClient.Get(urlStr)
	if e != nil {
		return err("search request failed: " + e.Error())
}

	defer resp.Body.Close()

	var result struct {
		AbstractText string   `json:"AbstractText"`
		RelatedTopics []struct {
			Text   string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	var parts []string
	if result.AbstractText != "" {
		parts = append(parts, result.AbstractText)

	for i, topic := range result.RelatedTopics {
		if i >= maxResults {
			break
		}
		parts = append(parts, fmt.Sprintf("- %s (%s)", topic.Text, topic.FirstURL))

	if len(parts) == 0 {
		return ok("No results found.")
}

	return ok(strings.Join(parts, "\n"))
}
}
}