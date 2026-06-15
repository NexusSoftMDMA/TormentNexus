package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func HandleGreeting(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello, " + name + "!")
}

func HandleFetchRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.github.com/search/repositories?q=topic:mcp-framework")
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse: " + e.Error())
}

	items, found := result["items"].([]interface{})
	if !found {
		return err("no items found")
}

	names := []string{}
	for _, item := range items {
		repo, found := item.(map[string]interface{})
		if !found {
			continue
		}
		name, found := repo["full_name"].(string)
		if found {
			names = append(names, name)

	}
	return ok("Repos: " + strings.Join(names, ", "))
}
}