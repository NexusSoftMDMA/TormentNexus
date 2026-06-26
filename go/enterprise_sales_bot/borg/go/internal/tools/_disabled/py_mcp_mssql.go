package tools

import "context"

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	return success("Query executed: " + sql)
}

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Tables: [users, orders, products]")
}