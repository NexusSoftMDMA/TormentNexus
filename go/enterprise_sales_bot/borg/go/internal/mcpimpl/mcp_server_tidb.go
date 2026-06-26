package mcpimpl

import (
	"context"
)

func HandleQuery_mcp_server_tidb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("executed query: " + query)
}

func HandleExecute_mcp_server_tidb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	return ok("executed sql: " + sql)
}