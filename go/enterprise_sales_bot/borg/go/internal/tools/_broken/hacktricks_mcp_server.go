package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type searchHackTricksResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type searchHackTricksResponse struct {
	Hits []searchHackTricksResult `json:"hits"`
}

func HandleSearchHackTricks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	searchURL := "https://book.hacktricks.xyz/search.json?q=" + url.QueryEscape(query)
	resp, e := http.DefaultClient.Get(searchURL)
	if e != nil {
		return err("failed to search: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("search returned status " + resp.Status)
}

	var data searchHackTricksResponse
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse search results: " + e.Error())
}

	if len(data.Hits) == 0 {
		return err("no results found")
}

	var result string
	for _, hit := range data.Hits {
		result += fmt.Sprintf("- [%s](%s)\n  %s\n\n", hit.Title, hit.URL, hit.Content)

	return ok(result)
}
}