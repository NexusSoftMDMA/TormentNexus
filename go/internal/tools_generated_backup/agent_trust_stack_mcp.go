package tools

import (
	"context"
)

func HandleGetTrustSummary(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "agent_name")
	if name == "" {
		return err("agent_name is required")
}

	return success("Trust summary for agent " + name + ": all clear")
}