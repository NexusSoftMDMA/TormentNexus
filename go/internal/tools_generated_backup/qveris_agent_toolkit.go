package tools

import (
	"context"
	"fmt"
)

func HandleListTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Available tools: tool1, tool2, tool3")
}

func HandleCallTool(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("parameter 'name' is required")
}

	return ok(fmt.Sprintf("Called tool: %s", name))
}