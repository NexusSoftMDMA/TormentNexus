package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetLatestNews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 5
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://cointelegraph.com/api/v1/news?limit=%d", limit), nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	list, found := result["data"].([]interface{})
	if !found {
		return err("unexpected response structure")
}

	var out string
	for i, item := range list {
		m, found := item.(map[string]interface{})
		if !found {
			continue
		}
		title, _ := m["title"].(string)
		url, _ := m["url"].(string)
		out += fmt.Sprintf("%d. %s (%s)\n", i+1, title, url)

	if out == "" {
		return success("no articles found")
}

	return ok(out)
}
}