package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	key := os.Getenv("GOOGLE_API_KEY")
	cx := os.Getenv("GOOGLE_CSE_ID")
	if key == "" || cx == "" {
		return err("GOOGLE_API_KEY and GOOGLE_CSE_ID must be set")
}

	apiURL := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s", url.QueryEscape(key), url.QueryEscape(cx), url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("search request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	items, found := result["items"].([]interface{})
	if !found {
		return ok("no results found")
}

	var snippets []string
	for _, item := range items {
		m, found := item.(map[string]interface{})
		if !found {
			continue
		}
		title, _ := m["title"].(string)
		link, _ := m["link"].(string)
		snippet, _ := m["snippet"].(string)
		snippets = append(snippets, fmt.Sprintf("Title: %s\nLink: %s\nSnippet: %s", title, link, snippet))

	return ok(fmt.Sprintf("Search results:\n%s", joinStrings(snippets, "\n---\n")))
}

}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}