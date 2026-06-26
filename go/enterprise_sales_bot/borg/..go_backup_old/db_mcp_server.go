package tools

import (
	"context"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query is required")
}

	return success("Executed query: " + q)
}

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, _ :=getString(args, "database")
	if db == "" {
		return err("database is required")
}

	return success("Tables in " + db + ": [users, orders, products]")
}