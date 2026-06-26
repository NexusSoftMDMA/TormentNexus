package mcpimpl

import (
	"context"
	"fmt"
)

func HandleListDatabases_aws_athena_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Databases: [\"default\"]")
}

func HandleRunQuery_aws_athena_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	return ok(fmt.Sprintf("Result for: %s", sql))
}