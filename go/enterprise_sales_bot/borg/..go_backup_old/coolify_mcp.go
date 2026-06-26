package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	token, _ :=getString(args, "api_token")
	if baseURL == "" || token == "" {
		return err("base_url and api_token are required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/v1/projects", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
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
		return err(fmt.Sprintf("API error (%d): %s", resp.StatusCode, string(body)))
	}
	return success(fmt.Sprintf("Projects: %s", string(body)))
}

func HandleListDeployments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	token, _ :=getString(args, "api_token")
	projectID, _ :=getString(args, "project_id")
	if baseURL == "" || token == "" || projectID == "" {
		return err("base_url, api_token, and project_id are required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/v1/projects/"+projectID+"/deployments", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
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
		return err(fmt.Sprintf("API error (%d): %s", resp.StatusCode, string(body)))
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json parse error: %v", e))
	}
	return ok(fmt.Sprintf("Deployments: %+v", result))
}