package tools

import "context"

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return success("Executed query: " + query + ", result: dummy")
}

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Tables: users, orders")
}