package mcpimpl

import (
	"context"
	"fmt"
)

func HandleAdd_calcnook_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	return success(fmt.Sprintf("Sum: %d", a+b))
}

func HandleMultiply_calcnook_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	return success(fmt.Sprintf("Product: %d", a*b))
}