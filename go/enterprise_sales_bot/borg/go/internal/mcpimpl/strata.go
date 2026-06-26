package mcpimpl

import "context"

func HandleHello_strata(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return success("Hello, " + name)
}

func HandlePing_strata(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}