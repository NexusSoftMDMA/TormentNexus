package tools

import "context"

var memories = make(map[string]string)

func HandleSetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" {
		return err("key is required")
}

	if value == "" {
		return err("value is required")
}

	memories[key] = value
	return success("memory stored")
}

func HandleGetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	value, found := memories[key]
	if !found {
		return err("memory not found")
}

	return ok(value)
}