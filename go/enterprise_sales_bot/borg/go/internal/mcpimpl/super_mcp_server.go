package mcpimpl

import (
	"context"
	"fmt"
)

func HandleSuper(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	msg := fmt.Sprintf("Hello, %s! Welcome to Super MCP Server.", name)
	return success(msg)
}

func HandleAdd_super_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	sum := a + b
	msg := fmt.Sprintf("Sum: %d", sum)
	return success(msg)
}