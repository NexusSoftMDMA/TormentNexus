package tools

import (
	"context"
	"fmt"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return ok("Echo: " + message)
}

func HandleAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	sum := a + b
	return ok(fmt.Sprintf("Sum: %d", sum))
}