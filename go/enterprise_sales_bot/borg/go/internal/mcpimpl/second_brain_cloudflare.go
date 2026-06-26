package mcpimpl

import (
	"context"
)

func HandleSave_second_brain_cloudflare(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	return ok("Saved: " + key + " = " + value)
}

func HandleLoad(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	return ok("Loaded: " + key)
}