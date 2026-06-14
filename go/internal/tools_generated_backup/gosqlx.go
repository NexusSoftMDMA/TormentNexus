package tools

import "context"

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("Executed query: " + query)
}