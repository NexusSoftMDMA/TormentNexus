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
	if baseURL == "" {
		baseURL = "https://api.comet.com/opik"
	}
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/projects", nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("JSON parse failed: %v", e))
}

	return ok(fmt.Sprintf("Projects: %v", result))
}

func HandleListExperiments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://api.comet.com/opik"
	}
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	projectID, _ :=getString(args, "project_id")
	if projectID == "" {
		return err("project_id is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/experiments?project_id="+projectID, nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("JSON parse failed: %v", e))
}

	return ok(fmt.Sprintf("Experiments: %v", result))
}