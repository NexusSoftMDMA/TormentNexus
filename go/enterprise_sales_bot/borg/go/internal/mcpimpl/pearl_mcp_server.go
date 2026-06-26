package mcpimpl

import (
	"context"
	"fmt"
)

func HandleEcho_pearl_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return success(msg)
}

func HandleAdd_pearl_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	sum := a + b
	return success(fmt.Sprintf("%d", sum))
}