package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleListProjects_mcp_model_context_protocol_projects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.example.com/mcp/projects")
	if e != nil {
		return err("failed to fetch projects: " + e.Error())
}

	defer resp.Body.Close()
	var projects []string
	if e := json.NewDecoder(resp.Body).Decode(&projects); e != nil {
		return err("decode error: " + e.Error())
}

	return ok(fmt.Sprintf("Projects: %s", strings.Join(projects, ", ")))
}

func HandleCreateProject_mcp_model_context_protocol_projects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("missing required 'name' argument")
}

	return success(fmt.Sprintf("Created project '%s'", name))
}