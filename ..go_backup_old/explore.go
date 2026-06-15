package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleExplore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		topic = "mcp-server"
	}
	url := fmt.Sprintf("https://api.github.com/search/repositories?q=topic:%s", topic)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()

	var result struct {
		Items []struct {
			FullName string `json:"full_name"`
		} `json:"items"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode JSON: " + e.Error())
}

	var names []string
	for _, item := range result.Items {
		names = append(names, item.FullName)

	if len(names) == 0 {
		return success("No repositories found for topic: " + topic)
}

	return success(fmt.Sprintf("Found %d repositories: %s", len(names), names))
}
}