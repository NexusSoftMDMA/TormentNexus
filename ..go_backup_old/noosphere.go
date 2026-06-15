package tools

import "context"

func HandleQuerySphere(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return success("Query received: " + query)
}