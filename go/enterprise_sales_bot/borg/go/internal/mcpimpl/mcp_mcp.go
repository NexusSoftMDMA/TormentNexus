package mcpimpl

import (
	"context"
)

func HandleEcho_mcp_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}

func HandleGreet_mcp_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	greeting := "Hello, " + name
	return ok(greeting)
}