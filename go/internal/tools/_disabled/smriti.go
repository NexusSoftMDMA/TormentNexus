package tools

import (
	"context"
	"fmt"
)

var memories = map[string]string{}

func HandleSetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" {
		return err("key is required")
}

	if value == "" {
		return err("value is required")
}

	memories[key] = value
	return ok(fmt.Sprintf("Memory set: %s", key))
}

func HandleGetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	value, found := memories[key]
	if !found {
		return err(fmt.Sprintf("Memory not found: %s", key))
}

	return success(value)
}