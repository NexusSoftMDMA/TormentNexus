package mcpimpl

import "context"

func HandleHello_mcp_agent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Hello from Mcp Agent")
}

func HandleEcho_mcp_agent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	return ok(input)
}