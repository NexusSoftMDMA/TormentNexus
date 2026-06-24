package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	apiToken, _ :=getString(args, "api_token")
	if baseURL == "" || apiToken == "" {
		return err("missing base_url or api_token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/v1/projects", nil)
	if e != nil {
		return err("request error: " + e.Error())
}

	req.Header.Set("X-API-Key", apiToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return ok(string(body))
}

func HandleListTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	apiToken, _ :=getString(args, "api_token")
	projectID, _ :=getString(args, "project_id")
	if baseURL == "" || apiToken == "" || projectID == "" {
		return err("missing base_url, api_token, or project_id")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/v1/projects/"+projectID+"/tasks", nil)
	if e != nil {
		return err("request error: " + e.Error())
}

	req.Header.Set("X-API-Key", apiToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return ok(string(body))
}