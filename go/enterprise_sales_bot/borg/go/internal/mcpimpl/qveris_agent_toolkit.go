package mcpimpl

import (
	"context"
	"fmt"
)

func HandleListTools_qveris_agent_toolkit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Available tools: tool1, tool2, tool3")
}

func HandleCallTool_qveris_agent_toolkit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("parameter 'name' is required")
}

	return ok(fmt.Sprintf("Called tool: %s", name))
}