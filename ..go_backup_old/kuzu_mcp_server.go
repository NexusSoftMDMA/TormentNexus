package tools

import "context"

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return ok("Query executed successfully: " + query)
}

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Tables: [users, products, orders]")
}