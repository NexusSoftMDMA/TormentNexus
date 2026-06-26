package mcpimpl

import (
	"context"
)

func HandleOctomind(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return ok("Processed: " + query)
}