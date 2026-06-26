package mcpimpl

import "context"

func HandleOpenProject_kicad_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	return ok("Opened project: " + project)
}

func HandleListProjects_kicad_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Listed projects")
}