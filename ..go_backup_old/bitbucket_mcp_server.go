package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListRepositories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workspace, _ :=getString(args, "workspace")
	if workspace == "" {
		return err("workspace is required")
}

	token, _ :=getString(args, "token")
	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s", workspace)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("creating request: %v", e))
}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http request: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("reading body: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json parse: %v", e))
}

	values, found := result["values"]
	if !found {
		return err("no values in response")
}

	return ok(fmt.Sprintf("Found repositories: %v", values))
}

}

func HandleGetRepository(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workspace, _ :=getString(args, "workspace")
	repo, _ :=getString(args, "repo")
	if workspace == "" || repo == "" {
		return err("workspace and repo are required")
}

	token, _ :=getString(args, "token")
	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", workspace, repo)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("creating request: %v", e))
}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http request: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("reading body: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json parse: %v", e))
}

	return ok(fmt.Sprintf("Repository info: %v", result))
}
}