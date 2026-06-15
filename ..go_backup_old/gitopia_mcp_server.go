package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGitopiaGetRepo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	if owner == "" || repo == "" {
		return err("owner and repo are required")
}

	url := fmt.Sprintf("https://api.gitopia.com/v1/repos/%s/%s", owner, repo)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success(fmt.Sprintf("Repo info: %+v", result))
}