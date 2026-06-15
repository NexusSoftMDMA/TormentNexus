package tools

import "context"

func HandleGetContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	return ok("Context value for " + key + " is: example_value")
}

func HandleSetContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	return success("Set context " + key + " to " + value)
}