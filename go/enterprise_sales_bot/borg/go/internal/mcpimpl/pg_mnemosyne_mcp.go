package mcpimpl

import (
	"context"
)

func HandlePing_pg_mnemosyne_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}

func HandleQuery_pg_mnemosyne_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("No query provided")
}

	_ = query
	return success("Query executed (mock)")
}