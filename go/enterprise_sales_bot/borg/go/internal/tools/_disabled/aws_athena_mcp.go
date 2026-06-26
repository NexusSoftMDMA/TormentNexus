package tools

import (
	"context"
	"fmt"
)

func HandleListDatabases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Databases: [\"default\"]")
}

func HandleRunQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	return ok(fmt.Sprintf("Result for: %s", sql))
}