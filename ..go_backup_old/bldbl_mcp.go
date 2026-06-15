package tools

import (
	"context"
	"fmt"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return ok(fmt.Sprintf("Echo: %s", msg))
}

func HandleAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	result := a + b
	return success(fmt.Sprintf("%d + %d = %d", a, b, result))
}