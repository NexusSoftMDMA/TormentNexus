package mcpimpl

import (
	"context"
	"fmt"
)

func HandleRegistepPing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong")
}

func HandleRegistepAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	result := a + b
	return ok(fmt.Sprintf("result: %d", result))
}