package mcpimpl

import (
	"context"
	"fmt"
)

func HandleHello_pfc_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Hello from Pfc MCP!")
}

func HandleAddNumbers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	sum := a + b
	return ok(fmt.Sprintf("Sum: %d", sum))
}