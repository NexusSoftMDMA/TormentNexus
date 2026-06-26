package mcpimpl

import (
	"context"
	"strings"
)

var memories_yourmemory = make(map[string][]string)

func HandleMemorySave(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	memories[key] = append(memories[key], value)
	return ok("saved")
}

func HandleMemoryRecall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	values, found := memories[key]
	if !found {
		return err("no memories found")
}

	return success("memories: " + strings.Join(values, ", "))
}