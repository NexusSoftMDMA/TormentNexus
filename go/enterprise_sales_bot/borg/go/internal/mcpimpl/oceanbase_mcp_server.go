package mcpimpl

import (
	"context"
	"fmt"
	"strings"
)

func HandleQuery_oceanbase_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	if sql == "" {
		return err("sql is required")
}

	result := fmt.Sprintf("Executed query: %s", sql)
	return ok(result)
}

func HandleListTables_oceanbase_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, _ :=getString(args, "database")
	if db == "" {
		db = "default"
	}
	tables := []string{"users", "orders", "products"}
	tableStr := strings.Join(tables, "\n")
	return success(fmt.Sprintf("Tables in database %s:\n%s", db, tableStr))
}