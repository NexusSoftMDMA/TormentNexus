package tools

import (
	"context"
	"fmt"
)

func HandleGetClarity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok(fmt.Sprintf("Clarity says hello, %s", name))
}

func HandleCount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	count, _ :=getInt(args, "count")
	return ok(fmt.Sprintf("Count is %d", count))
}