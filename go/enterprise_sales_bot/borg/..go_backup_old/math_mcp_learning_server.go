package tools

import (
	"context"
	"fmt"
)

func HandleAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	return success(fmt.Sprintf("Result: %d", a+b))
}

func HandleMultiply(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	return success(fmt.Sprintf("Result: %d", a*b))
}