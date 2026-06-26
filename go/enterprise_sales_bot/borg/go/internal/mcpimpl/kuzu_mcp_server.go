package mcpimpl

import "context"

func HandleQuery_kuzu_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return ok("Query executed successfully: " + query)
}

func HandleListTables_kuzu_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Tables: [users, products, orders]")
}