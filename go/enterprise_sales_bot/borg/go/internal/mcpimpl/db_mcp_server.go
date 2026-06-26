package mcpimpl

import (
	"context"
)

func HandleQuery_db_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query is required")
}

	return success("Executed query: " + q)
}

func HandleListTables_db_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, _ :=getString(args, "database")
	if db == "" {
		return err("database is required")
}

	return success("Tables in " + db + ": [users, orders, products]")
}