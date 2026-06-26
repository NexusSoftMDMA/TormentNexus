package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGreet_gm_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	msg := fmt.Sprintf("Hello, %s! Welcome to Gm MCP.", name)
	return ok(msg)
}

func HandleEcho_gm_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	return ok(text)
}