package mcpimpl

import "context"

func HandleRabel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Rabel"
	}
	return ok("Hello from " + name + "!")
}