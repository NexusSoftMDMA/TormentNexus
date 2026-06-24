package tools

import (
	"context"
	"fmt"
)

func HandleFizzyEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(fmt.Sprintf("Fizzy says: %s", msg))
}

func HandleFizzyAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	return success(fmt.Sprintf("%d + %d = %d", a, b, a+b))
}