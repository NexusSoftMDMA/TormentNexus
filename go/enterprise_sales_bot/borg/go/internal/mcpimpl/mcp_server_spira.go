package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

func HandleListSpiraProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := os.Getenv("SPIRA_BASE_URL")
	apiKey := os.Getenv("SPIRA_API_KEY")
	if baseURL == "" || apiKey == "" {
		return err("SPIRA_BASE_URL and SPIRA_API_KEY must be set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/projects", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	jsonBytes, _ := json.Marshal(result)
	return ok(string(jsonBytes))
}

func HandleGetSpiraProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getInt(args, "project_id")
	if projectID == 0 {
		return err("project_id is required")
}

	baseURL := os.Getenv("SPIRA_BASE_URL")
	apiKey := os.Getenv("SPIRA_API_KEY")
	if baseURL == "" || apiKey == "" {
		return err("SPIRA_BASE_URL and SPIRA_API_KEY must be set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/projects/"+strconv.Itoa(projectID), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	jsonBytes, _ := json.Marshal(result)
	return ok(string(jsonBytes))
}