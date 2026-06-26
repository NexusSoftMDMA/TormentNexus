package mcpimpl

import (
	"context"
	"fmt"
)

func HandleSearch_rendex_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return ok(fmt.Sprintf("Search results for '%s'", query))
}