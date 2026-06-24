package tools

import "context"

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("users, orders, products")
}

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	if sql == "" {
		return err("query is empty")
}

	return ok("Query executed successfully: " + sql)
}