package mcpimpl

import "context"

var memories_mcp_memory_service = make(map[string]string)

func HandleSetMemory_mcp_memory_service(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	memories[key] = value
	return ok("Memory stored")
}

func HandleGetMemory_mcp_memory_service(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, found := memories[key]
	if !found {
		return err("Memory not found")
}

	return ok(value)
}