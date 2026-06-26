package mcpimpl

import (
	"context"
	"fmt"
)

func HandleEcho_orizn_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return ok(fmt.Sprintf("Echo: %s", message))
}

func HandleAdd_orizn_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	return ok(fmt.Sprintf("Sum: %d", a+b))
}