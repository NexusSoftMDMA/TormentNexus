package mcpimpl

import "context"

func HandleGetInfo_mcp_server_tibet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Tibet"
	}
	msg := "Hello from " + name + "! MCP Server Tibet is running."
	return ok(msg)
}