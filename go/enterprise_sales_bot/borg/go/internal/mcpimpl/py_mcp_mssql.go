package mcpimpl

import "context"

func HandleQuery_py_mcp_mssql(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	return success("Query executed: " + sql)
}

func HandleListTables_py_mcp_mssql(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Tables: [users, orders, products]")
}