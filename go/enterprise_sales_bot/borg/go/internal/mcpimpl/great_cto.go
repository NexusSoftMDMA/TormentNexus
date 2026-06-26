package mcpimpl

import (
	"context"
)

func HandleGreatCto(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	stack, _ :=getString(args, "stack")
	path, _ :=getString(args, "path")
	msg := "Installed great_cto plugin"
	if stack != "" {
		msg += " for " + stack
	}
	if path != "" {
		msg += " at " + path
	}
	return ok(msg)
}