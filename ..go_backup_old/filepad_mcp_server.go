package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListWorkspaces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	resp, e := http.DefaultClient.Get(baseURL + "/workspaces")
	if e != nil {
		return err(fmt.Sprintf("failed to call API: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Workspaces: %s", string(body)))
}

func HandleGetWorkspace(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	id, _ :=getString(args, "workspace_id")
	if id == "" {
		return err("workspace_id is required")
}

	resp, e := http.DefaultClient.Get(baseURL + "/workspaces/" + id)
	if e != nil {
		return err(fmt.Sprintf("failed to call API: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Workspace: %s", string(body)))
}