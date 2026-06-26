package mcpimpl

import (
	"context"
)

func HandleSearch_ragstack_lambda(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return success("Search result for: " + query)
}

func HandleQuery_ragstack_lambda(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return success("Query result for: " + query)
}