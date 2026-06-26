package mcpimpl

import "context"

func HandleListStudios(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("List of studios: Studio1, Studio2")
}

func HandleGetStudio(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("Studio: " + name)
}