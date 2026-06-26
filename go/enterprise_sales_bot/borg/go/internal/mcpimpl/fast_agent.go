package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGreet_fast_agent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Hello from Fast Agent MCP server!")
}

func HandleEcho_fast_agent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return success(fmt.Sprintf("Echo: %s", msg))
}