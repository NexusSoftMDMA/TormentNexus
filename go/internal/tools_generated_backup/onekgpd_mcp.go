package tools

import "context"

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query parameter required")
}

	return ok("Query received: " + q)
}

func HandleList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Available tools: query, list")
}