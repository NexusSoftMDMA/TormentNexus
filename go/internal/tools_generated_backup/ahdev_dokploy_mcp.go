package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func HandleGetProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	url := fmt.Sprintf("%s/api/projects", baseURL)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Projects: %v", result))
}

func HandleGetDeployments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	projectID, _ :=getString(args, "project_id")
	limit, _ :=getInt(args, "limit")
	u := fmt.Sprintf("%s/api/deployments?project_id=%s", baseURL, projectID)
	if limit > 0 {
		u += "&limit=" + strconv.Itoa(limit)

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Deployments: %v", result))
}
}