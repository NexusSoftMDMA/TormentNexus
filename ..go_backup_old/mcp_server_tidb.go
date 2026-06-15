package tools

import (
	"context"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("executed query: " + query)
}

func HandleExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	return ok("executed sql: " + sql)
}