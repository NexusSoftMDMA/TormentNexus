package tools

import (
	"context"
	"fmt"
)

func HandleRedisGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("missing key")
}

	return ok(fmt.Sprintf("GET %s: (mock) value", key))
}

func HandleRedisSet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	val, _ :=getString(args, "value")
	if key == "" {
		return err("missing key")
}

	return success(fmt.Sprintf("SET %s %s", key, val))
}// touch 1781132132
