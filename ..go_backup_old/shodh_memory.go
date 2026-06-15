package tools

import (
	"context"
)

var memoryStore = make(map[string]string)

func HandleStoreMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	memoryStore[key] = value
	return ok("Stored memory for key: " + key)
}

func HandleGetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, found := memoryStore[key]
	if !found {
		return err("Memory not found for key: " + key)
	}
	return success(value)
}