package tools

import "context"

func HandleHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Hello from Squad Mcp")
}