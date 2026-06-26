package tools

import "context"

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dbName, _ :=getString(args, "database")
	if dbName == "" {
		return err("missing database parameter")
}

	return ok("Tables in " + dbName + ": users, orders, products")
}

func HandleExecuteQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query parameter")
}

	return success("Query executed: " + query)
}