package tools

import (
	"context"
	"net/http"
)

func HandleExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	_ = http.DefaultClient
	return ok("executed: " + sql)
}

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	_ = http.DefaultClient
	return ok("query result for: " + sql)
}