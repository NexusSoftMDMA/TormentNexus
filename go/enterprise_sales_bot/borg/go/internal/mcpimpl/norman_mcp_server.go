package mcpimpl

import (
	"context"
	"fmt"
)

func HandleHello_norman_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok(fmt.Sprintf("Hello, %s! From Norman Mcp Server.", name))
}

func HandleAdd_norman_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	sum := a + b
	return ok(fmt.Sprintf("Sum of %d and %d is %d.", a, b, sum))
}