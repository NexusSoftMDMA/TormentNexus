package mcpimpl

import (
	"context"
	"fmt"
)

func HandleEcho_edwin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return ok(fmt.Sprintf("Echo: %s", msg))
}

func HandleAdd_edwin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	sum := a + b
	return ok(fmt.Sprintf("Sum: %d", sum))
}