//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HandleListRepos returns a list of repositories for a given user
func HandleListRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	user, _ :=getString(args, "user")
	if user == "" {
		return err("user is required")
}

	url := fmt.Sprintf("https://api.github.com/users/%s/repos", user)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch repos")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var repos []map[string]interface{}
	if e := json.Unmarshal(body, &repos); e != nil {
		return err("failed to parse repos")
}

	names := make([]string, 0, len(repos))
	for _, r := range repos {
		if name, found := r["name"].(string); found {
			names = append(names, name)

	}
	return ok(fmt.Sprintf("Repos: %v", names))
}

}

// HandleGetRepoContent fetches a file from a repository
func HandleGetRepoContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	path, _ :=getString(args, "path")
	if owner == "" || repo == "" || path == "" {
		return err("owner, repo, and path are required")
}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch content")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return success(string(body))
}
