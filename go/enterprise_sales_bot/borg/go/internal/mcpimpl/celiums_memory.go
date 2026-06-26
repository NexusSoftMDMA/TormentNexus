package mcpimpl

import (
	"context"
)

func HandleStoreMemory_celiums_memory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" || value == "" {
		return err("key and value are required")
}

	// In a real implementation, here you would store the memory.
	return success("memory stored: " + key)
}

func HandleRecallMemory_celiums_memory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	// In a real implementation, here you would retrieve the memory.
	return ok("memory for " + key + ": sample value")
}