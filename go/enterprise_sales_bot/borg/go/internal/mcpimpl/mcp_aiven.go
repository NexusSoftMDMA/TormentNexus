package mcpimpl

import "context"

func HandleListProjects_mcp_aiven(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Aiven projects listed")
}

func HandleGetProject_mcp_aiven(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "project_name")
	return success("Aiven project: " + name)
}