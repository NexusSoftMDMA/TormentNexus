package mcpimpl

import "context"

func HandleRiskstate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	state, _ :=getString(args, "state")
	if state == "" {
		return err("state is required")
}

	return success("Risk state: " + state)
}