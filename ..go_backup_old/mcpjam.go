package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleSearchRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		topic = "mcp-tools"
	}
	url := "https://api.github.com/search/repositories?q=topic:" + topic
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("User-Agent", "Mcpjam")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("GitHub API error: " + resp.Status)
}

	var result struct {
		Items []struct {
			Name string `json:"name"`
		} `json:"items"`
	}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err(e.Error())
}

	var repos []string
	for _, item := range result.Items {
		repos = append(repos, item.Name)

	return ok("Found repos: " + strings.Join(repos, ", "))
}
}