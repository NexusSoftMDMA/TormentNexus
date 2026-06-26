package mcpimpl

import (
	"context"
)

func HandleSearch_jetbrains_index_mcp_plugin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return success("found results for " + query)
}

func HandleInfo_jetbrains_index_mcp_plugin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	return ok("info for " + id)
}