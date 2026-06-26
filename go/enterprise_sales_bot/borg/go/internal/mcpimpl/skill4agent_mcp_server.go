package mcpimpl

import "context"

func HandleGreet_skill4agent_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok("Hello " + name)
}

func HandleEcho_skill4agent_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}