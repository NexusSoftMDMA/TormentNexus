package tools

import (
	"context"
	"strconv"
)

func HandleAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	return ok("Result: " + strconv.Itoa(a+b))
}