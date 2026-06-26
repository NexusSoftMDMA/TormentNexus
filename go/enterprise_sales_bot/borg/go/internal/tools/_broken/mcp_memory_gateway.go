package mcpimpl

import (
	"context"
	"fmt"
)

var memoryStore_mcp_memory_gateway = make(map[string]string)

func HandleSaveMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	memoryStore[key] = value
	return ok(fmt.Sprintf("Saved memory for key '%s'", key))
}

func HandleGetMemory_mcp_memory_gateway(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, found := memoryStore[key]
	if !found {
		return err("Memory not found")
}

	return ok(value)
}