package mcpimpl

import (
	"context"
)

func HandleQuery_dexto(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return ok("You said: " + query)
}

func HandlePing_dexto(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}