package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchBooks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := fmt.Sprintf("https://openlibrary.org/search.json?q=%s", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("API returned status " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e = json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	docs, found := data["docs"].([]interface{})
	if !found || len(docs) == 0 {
		return ok("No results found for: " + query)
}

	result := fmt.Sprintf("Search results for '%s':\n", query)
	for i, raw := range docs {
		if i >= 5 {
			break
		}
		doc, _ := raw.(map[string]interface{})
		title, _ := doc["title"].(string)
		authors, _ := doc["author_name"].([]interface{})
		author := ""
		if len(authors) > 0 {
			author, _ = authors[0].(string)

		result += fmt.Sprintf("%d. %s by %s\n", i+1, title, author)

	return ok(result)
}
}
}