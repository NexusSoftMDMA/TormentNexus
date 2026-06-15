//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	max, _ :=getInt(args, "max_results")
	if max <= 0 || max > 20 {
		max = 5
	}
	u := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json", url.QueryEscape(query))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(e.Error())
}

	results := []map[string]string{}
	if raw, found := data["RelatedTopics"].([]interface{}); found {
		for i, r := range raw {
			if i >= max {
				break
			}
			if m, found := r.(map[string]interface{}); found {
				text, _ := m["Text"].(string)
				url, _ := m["FirstURL"].(string)
				results = append(results, map[string]string{"text": text, "url": url})

		}
	}
	out := ""
	for _, r := range results {
		out += fmt.Sprintf("- %s (%s)\n", r["text"], r["url"])

	if out == "" {
		out = "No results found."
	}
	return ok(out)
}
}
}
