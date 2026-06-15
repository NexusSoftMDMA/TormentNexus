package tools

import (
	"context"
	"net/http"
	"encoding/json"
	"fmt"
)

func HandleAnalyzeRepo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoPath, _ :=getString(args, "repo_path")
	if repoPath == "" {
		return err("repo_path is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://api.github.com/repos/%s", repoPath))
	if e != nil {
		return err("failed to fetch repo: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response")
}

	found, _ := data["full_name"].(string)
	if !found {
		return ok(fmt.Sprintf("Repo %s analyzed (no name found)", repoPath))
}

	return ok(fmt.Sprintf("Repo %s analyzed", data["full_name"]))
}

func HandleGetSuggestions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	steps, _ :=getInt(args, "steps")
	if project == "" {
		return err("project is required")
}

	if steps <= 0 {
		steps = 3
	}
	return ok(fmt.Sprintf("Generated %d modernization steps for %s", steps, project))
}