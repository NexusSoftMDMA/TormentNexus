package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSearchMcpPlugins(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		query = "topic:mcp-plugin"
	}
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	url := fmt.Sprintf("https://api.github.com/search/repositories?q=%s&per_page=%d", query, limit)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch repositories: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Items []struct {
			FullName string `json:"full_name"`
		} `json:"items"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	var names []string
	for _, item := range result.Items {
		names = append(names, item.FullName)

	message := fmt.Sprintf("Found %d repositories:\n", len(names))
	for _, n := range names {
		message += n + "\n"
	}
	return ok(message)
}
}