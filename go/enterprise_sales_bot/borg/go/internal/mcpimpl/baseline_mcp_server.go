package mcpimpl

import (
	"context"
)

func HandleEcho_baseline_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok("Echo: " + msg)
}

func HandleAdd_baseline_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	result := a + b
	return ok("Result: " + string(rune(result)))
}