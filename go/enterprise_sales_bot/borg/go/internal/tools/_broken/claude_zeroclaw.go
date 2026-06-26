package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleSearchRepos_claude_zeroclaw(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	http.DefaultClient = http.DefaultClient
	q, _ :=getString(args, "topic")
	if q == "" {
		q = "claude-mcp"
	}
	url := fmt.Sprintf("https://api.github.com/search/repositories?q=topic:%s", q)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("GitHub API returned " + resp.Status)
}

	var result struct {
		Items []struct {
			FullName string `json:"full_name"`
			HTMLURL  string `json:"html_url"`
			Description string `json:"description"`
		} `json:"items"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error")
}

	if len(result.Items) == 0 {
		return success("No repositories found")
}

	var b strings.Builder
	for _, item := range result.Items {
		b.WriteString(fmt.Sprintf("- %s: %s\n  %s\n", item.FullName, item.HTMLURL, item.Description))

	return ok(b.String())
}

}

func HandleGetRepo_claude_zeroclaw(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	if owner == "" || repo == "" {
		return err("owner and repo required")
}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("GitHub API returned " + resp.Status)
}

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error")
}

	name, _ := data["full_name"].(string)
	desc, _ := data["description"].(string)
	stars := int(data["stargazers_count"].(float64))
	return ok(fmt.Sprintf("Repo: %s\nStars: %d\nDescription: %s", name, stars, desc))
}