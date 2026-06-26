package mcpimpl

import "context"

func HandleGetStatus_edgedelta_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Edgedelta server is running")
}

func HandleGetVersion_edgedelta_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Edgedelta MCP Server v1.0.0")
}