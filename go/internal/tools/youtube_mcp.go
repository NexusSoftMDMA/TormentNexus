//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchVideos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey is required")
}

	u, e := url.Parse("https://www.googleapis.com/youtube/v3/search")
	if e != nil {
		return err("failed to parse URL")
}

	q := u.Query()
	q.Set("part", "snippet")
	q.Set("q", query)
	q.Set("key", apiKey)
	q.Set("maxResults", "10")
	u.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON")
}

	items, found := result["items"].([]interface{})
	if !found {
		return ok("no results")
}

	out := ""
	for _, item := range items {
		if m, found := item.(map[string]interface{}); found {
			snippet, _ := m["snippet"].(map[string]interface{})
			title, _ := snippet["title"].(string)


-reasoner (deepseek)*,
},
},
}
}