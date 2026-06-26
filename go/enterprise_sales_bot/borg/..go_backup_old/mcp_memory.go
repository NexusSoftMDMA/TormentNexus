package tools

import "context"

var memory = map[string]string{}

func HandleRemember(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	memory[key] = value
	return ok("remembered")
}

func HandleRecall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, found := memory[key]
	if !found {
		return err("key not found")
}

	return success(value)
}