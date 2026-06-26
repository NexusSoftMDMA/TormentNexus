package mcpimpl

import "context"

func HandleCreateMcpServer_mcp_server_creator(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("MCP server '" + name + "' created successfully")
}