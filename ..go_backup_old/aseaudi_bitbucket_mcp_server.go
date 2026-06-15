package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workspace, _ :=getString(args, "workspace")
	if workspace == "" {
		return err("workspace is required")
}

	token, _ :=getString(args, "token")
	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s", workspace)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("JSON parse error: %v", e))
}

	return ok(fmt.Sprintf("Repositories: %+v", result))
}

}

func HandleListPullRequests(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workspace, _ :=getString(args, "workspace")
	repo, _ :=getString(args, "repo")
	if workspace == "" || repo == "" {
		return err("workspace and repo are required")
}

	token, _ :=getString(args, "token")
	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests", workspace, repo)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("JSON parse error: %v", e))
}

	return ok(fmt.Sprintf("Pull requests: %+v", result))
}
}