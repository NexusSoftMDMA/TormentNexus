package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListRepos_github_repos_manager_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("missing GitHub token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/repos?per_page=100", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var repos []map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&repos); e != nil {
		return err("decode failed: " + e.Error())
}

	var names []string
	for _, r := range repos {
		names = append(names, r["name"].(string))

	return ok(fmt.Sprintf("Repositories: %v", names))
}

}

func HandleCreateRepo_github_repos_manager_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("missing GitHub token")
}

	name, _ :=getString(args, "name")
	if name == "" {
		return err("repository name is required")
}

	body := map[string]interface{}{"name": name}
	if desc := getString(args, "description"); desc != "" {
		body["description"] = desc
	}
	payload, e := json.Marshal(body)
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.github.com/user/repos", bytes.NewBuffer(payload))
	if e != nil {
		return err("create request failed: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	if url, found := result["html_url"]; found {
		return success(fmt.Sprintf("Created repository: %s", url.(string)))
}

	return err("unexpected response from GitHub")
}