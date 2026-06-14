package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		topic = "mcp-integration"
	}
	url := "https://api.github.com/search/repositories?q=topic:" + topic + "&per_page=10"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to query GitHub: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	items, found := data["items"].([]interface{})
	if !found {
		return err("no items found")
}

	names := make([]string, 0, len(items))
	for _, item := range items {
		m, f := item.(map[string]interface{})
		if !f {
			continue
		}
		name, _ := m["full_name"].(string)
		if name != "" {
			names = append(names, name)

	}
	return ok(fmt.Sprintf("Found %d repos: %v", len(names), names))
}

}

func HandleGetRepo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	if owner == "" || repo == "" {
		return err("owner and repo are required")
}

	url := "https://api.github.com/repos/" + owner + "/" + repo
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to query GitHub: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	desc, _ := data["description"].(string)
	result := fmt.Sprintf("Repo: %s/%s - %s", owner, repo, desc)
	return success(result)
}