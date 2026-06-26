package mcpimpl

import "context"

func HandleHello_squad_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Hello from Squad Mcp")
}