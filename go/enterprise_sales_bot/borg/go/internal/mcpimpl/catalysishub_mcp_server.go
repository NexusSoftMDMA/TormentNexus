package mcpimpl

import "context"

func HandleGetInfo_catalysishub_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Catalysishub"
	}
	return success("Hello from " + name + " MCP Server!")
}