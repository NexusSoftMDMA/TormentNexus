package mcpimpl

import (
	"context"
	"net/http"
)

func HandleExecute_yukihito_mysql_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	_ = http.DefaultClient
	return ok("executed: " + sql)
}

func HandleQuery_yukihito_mysql_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	_ = http.DefaultClient
	return ok("query result for: " + sql)
}