package mcpimpl

import (
	"context"
)

func HandleLspaceInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	return success("Lspace result for '" + query + "': The answer is 42.")
}