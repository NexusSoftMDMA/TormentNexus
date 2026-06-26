package mcpimpl

import "context"

func HandleHello_boostspace_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	msg := "Hello " + name + " from Boostspace MCP Server!"
	return ok(msg)
}

func HandleEcho_boostspace_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		msg = "No message provided"
	}
	return ok(msg)
}