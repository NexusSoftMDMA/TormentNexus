package mcpimpl

import (
	"context"
)

func HandleContextMode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "context_name")
	if name == "" {
		return err("context_name is required")
}

	return ok("Context mode: " + name)
}