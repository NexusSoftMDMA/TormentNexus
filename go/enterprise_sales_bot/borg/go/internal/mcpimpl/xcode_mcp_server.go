package mcpimpl

import "context"

func HandleOpenProject_xcode_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	return success("opened project at " + path)
}

func HandleListSchemes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectPath, _ :=getString(args, "projectPath")
	if projectPath == "" {
		return err("projectPath is required")
}

	return ok("schemes: [Debug, Release]")
}