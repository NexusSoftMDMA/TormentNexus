//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleListWorkspaces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("ASANA_ACCESS_TOKEN")
	if token == "" {
		return err("ASANA_ACCESS_TOKEN not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://app.asana.com/api/1.0/workspaces", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	data, _ := json.Marshal(result["data"])
	return ok(string(data))
}

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("ASANA_ACCESS_TOKEN")
	if token == "" {
		return err("ASANA_ACCESS_TOKEN not set")
}

	workspace, _ :=getString(args, "workspace")
	url := "https://app.asana.com/api/1.0/projects"
	if workspace != "" {
		url += "?workspace=" + workspace
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	data, _ := json.Marshal(result["data"])
	return ok(string(data))
}
