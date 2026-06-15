package tools

import "context"

func HandleQuerywiseQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("Query executed: " + query)
}