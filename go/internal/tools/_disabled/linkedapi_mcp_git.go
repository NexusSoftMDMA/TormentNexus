package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleSearchRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	url := fmt.Sprintf("https://api.github.com/search/repositories?q=%s", query)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	items, found := result["items"].([]interface{})
	if !found {
		return ok("no repositories found")
}

	var sb strings.Builder
	for i, item := range items {
		repo, _ := item.(map[string]interface{})
		name, _ := repo["full_name"].(string)
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, name))

	return ok(sb.String())
}

}

func HandleGetUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username parameter is required")
}

	url := fmt.Sprintf("https://api.github.com/users/%s", username)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var user map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&user); e != nil {
		return err("decode error: " + e.Error())
}

	login, _ := user["login"].(string)
	name, _ := user["name"].(string)
	return ok(fmt.Sprintf("User: %s (%s)", login, name))
}