package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleScaffold(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectName, _ :=getString(args, "projectName")
	if projectName == "" {
		return err("projectName is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://pwa-kit.example.com/scaffold/%s", projectName), nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	return ok(fmt.Sprintf("Scaffolded project %s", projectName))
}

func HandleBuild_salesforce_pwa_kit_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectName, _ :=getString(args, "projectName")
	if projectName == "" {
		return err("projectName is required")
}

	return success(fmt.Sprintf("Build initiated for %s", projectName))
}