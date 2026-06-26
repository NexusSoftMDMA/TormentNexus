package mcpimpl

import "context"

func HandleListTables_mcp_server_mariadb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("users, orders, products")
}

func HandleQuery_mcp_server_mariadb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	if sql == "" {
		return err("query is empty")
}

	return ok("Query executed successfully: " + sql)
}