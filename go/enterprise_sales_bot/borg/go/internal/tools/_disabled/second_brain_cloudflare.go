package tools

import (
	"context"
)

func HandleSave(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	return ok("Saved: " + key + " = " + value)
}

func HandleLoad(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	return ok("Loaded: " + key)
}