package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("missing token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://connection.keboola.com/v2/projects", nil)
	if e != nil {
		return err(fmt.Sprintf("request creation: %v", e))
}

	req.Header.Set("X-StorageApi-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned %d", resp.StatusCode))
}

	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode: %v", e))
}

	return ok(fmt.Sprintf("Projects: %v", result))
}

func HandleGetProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("missing token")
}

	projectID, _ :=getString(args, "projectId")
	if projectID == "" {
		return err("missing projectId")
}

	url := fmt.Sprintf("https://connection.keboola.com/v2/projects/%s", projectID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("request creation: %v", e))
}

	req.Header.Set("X-StorageApi-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned %d", resp.StatusCode))
}

	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode: %v", e))
}

	return ok(fmt.Sprintf("Project: %v", result))
}