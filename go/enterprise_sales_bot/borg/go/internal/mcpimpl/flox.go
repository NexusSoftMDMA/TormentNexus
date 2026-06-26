package mcpimpl

import (
	"context"
)

func HandleFloxSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	_ = query
	return ok("Flox search received")
}