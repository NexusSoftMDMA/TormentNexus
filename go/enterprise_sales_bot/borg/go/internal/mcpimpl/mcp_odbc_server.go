package mcpimpl

import (
	"context"
)

func HandleExecuteQuery_mcp_odbc_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	connStr, _ :=getString(args, "connection_string")
	query, _ :=getString(args, "sql")
	return ok("executed query: " + query + " on " + connStr)
}