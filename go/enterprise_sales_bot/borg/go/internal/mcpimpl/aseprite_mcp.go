package mcpimpl

import "context"

func HandleGetActiveSprite(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "default"
	}
	return success("Active sprite: " + name)
}