package mcpimpl

import (
	"context"
	"strconv"
)

func HandleAdd_primitiv(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	return ok("Result: " + strconv.Itoa(a+b))
}