package mcpimpl

import "context"

func HandleFindAll(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	return ok("Found all for model: " + model)
}

func HandleFindOne(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	id, _ :=getString(args, "id")
	return ok("Found one for model " + model + " with id " + id)
}