package tools

import "context"

var memories = map[string]string{}

func HandleSetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	memories[key] = value
	return ok("memory stored")
}

func HandleGetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, found := memories[key]
	if !found {
		return err("key not found")
}

	return success(value)
}