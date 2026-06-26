package mcpimpl

import "context"

func HandleQuery_mysql_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return success("Executed query: " + query + ", result: dummy")
}

func HandleListTables_mysql_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Tables: users, orders")
}