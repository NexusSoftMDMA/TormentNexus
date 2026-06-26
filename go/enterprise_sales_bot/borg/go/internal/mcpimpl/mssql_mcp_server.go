package mcpimpl

import "context"

func HandleQuery_mssql_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	_ = ctx
	return ok("Query executed: " + query)
}