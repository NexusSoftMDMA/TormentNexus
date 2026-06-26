package mcpimpl

import "context"

func HandleQuery_sage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("result: " + query)
}