package tools

import "context"

var memories = make(map[string]string)

func HandleSetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	memories[key] = value
	return ok("Memory stored")
}

func HandleGetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, found := memories[key]
	if !found {
		return err("Memory not found")
}

	return ok(value)
}