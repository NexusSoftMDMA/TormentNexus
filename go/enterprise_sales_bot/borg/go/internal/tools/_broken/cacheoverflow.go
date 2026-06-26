package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchStackOverflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter required")
}

	u := fmt.Sprintf("https://api.stackexchange.com/2.3/search?order=desc&sort=activity&intitle=%s&site=stackoverflow", url.QueryEscape(query))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Items []struct {
			Title string `json:"title"`
			Link  string `json:"link"`
		} `json:"items"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	if len(result.Items) == 0 {
		return ok("No results found")
}

	msg := "Top Stack Overflow results:\n"
	for i, item := range result.Items {
		if i >= 5 {
			break
		}
		msg += fmt.Sprintf("%d. %s\n   %s\n", i+1, item.Title, item.Link)

	return ok(msg)
}
}