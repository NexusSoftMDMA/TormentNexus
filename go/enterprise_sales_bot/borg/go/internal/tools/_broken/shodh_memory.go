package mcpimpl

import (
	"context"
)

var memoryStore_shodh_memory = make(map[string]string)

func HandleStoreMemory_shodh_memory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	memoryStore[key] = value
	return ok("Stored memory for key: " + key)
}

func HandleGetMemory_shodh_memory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, found := memoryStore[key]
	if !found {
		return err("Memory not found for key: " + key)
	}
	return success(value)
}