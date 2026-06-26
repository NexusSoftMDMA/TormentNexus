package mcpimpl

import (
	"context"
)

func HandleQuery_mcp_vertica(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return success("Executed: " + query)
}

func HandlePing_mcp_vertica(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}