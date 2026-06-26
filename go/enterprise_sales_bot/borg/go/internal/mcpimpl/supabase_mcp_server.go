package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleExecuteQuery_supabase_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	projectRef, _ :=getString(args, "project_ref")
	apiKey, _ :=getString(args, "api_key")
	if query == "" || projectRef == "" || apiKey == "" {
		return err("missing required parameters: query, project_ref, api_key")
}

	url := fmt.Sprintf("https://api.supabase.com/v1/projects/%s/database/query", projectRef)
	body, _ := json.Marshal(map[string]string{"query": query})
	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	return ok("query executed successfully")
}

func HandleListProjects_supabase_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("missing api_key parameter")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.supabase.com/v1/projects", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	return ok("projects listed successfully")
}