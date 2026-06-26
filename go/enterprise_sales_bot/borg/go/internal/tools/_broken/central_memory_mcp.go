package mcpimpl

import (
	"context"
	"sync"
)

var memory_central_memory_mcp sync.Map

func HandleRemember_central_memory_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	memory.Store(key, value)
	return ok("stored")
}

func HandleRecall_central_memory_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, found := memory.Load(key)
	if !found {
		return err("not found")
}

	return success(value.(string))
}