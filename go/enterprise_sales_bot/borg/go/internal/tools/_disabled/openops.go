package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ghRepo struct {
	Description string `json:"description"`
}

func HandleGetRepo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	if owner == "" || repo == "" {
		return err("owner and repo are required")
}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("GitHub API returned status " + resp.Status)
}

	var repoData ghRepo
	if e := json.NewDecoder(resp.Body).Decode(&repoData); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok("Description: " + repoData.Description)
}