package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListRepos_hasmcp_ce(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		topic = "mcp-framework"
	}
	limit, _ :=getInt(args, "limit")
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	url := fmt.Sprintf("https://api.github.com/search/repositories?q=topic:%s&per_page=%d&sort=stars&order=desc", topic, limit)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		Items []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"items"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if len(result.Items) == 0 {
		return ok("No repositories found for topic: " + topic)
}

	out := fmt.Sprintf("Top %d repositories for topic '%s':\n", limit, topic)
	for _, item := range result.Items {
		out += fmt.Sprintf("- %s: %s\n", item.Name, item.Description)

	return success(out)
}
}