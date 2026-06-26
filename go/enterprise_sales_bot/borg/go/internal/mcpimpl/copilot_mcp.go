package mcpimpl

import (
	"context"
)

func HandleGetCopilotInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	feature, _ :=getString(args, "feature")
	if feature == "" {
		feature = "general"
	}
	return success("Copilot Mcp feature: " + feature)
}

func HandleCopilotQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return ok("Copilot query processed: " + query)
}