package tools

import (
	"context"
	"fmt"
)

func HandleGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	return ok(fmt.Sprintf("value_for_%s", key))
}

func HandleSet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	return ok(fmt.Sprintf("set %s = %s", key, value))
}