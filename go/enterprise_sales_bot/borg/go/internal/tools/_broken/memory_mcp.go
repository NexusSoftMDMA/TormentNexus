package mcpimpl

import "context"

var memories_memory_mcp = map[string]string{}

func HandleSetMemory_memory_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	memories[key] = value
	return ok("memory stored")
}

func HandleGetMemory_memory_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, found := memories[key]
	if !found {
		return err("key not found")
}

	return success(value)
}