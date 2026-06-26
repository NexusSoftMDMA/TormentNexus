package mcpimpl

import "context"

func HandleExecuteSelect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("Executing SELECT query: " + query)
}