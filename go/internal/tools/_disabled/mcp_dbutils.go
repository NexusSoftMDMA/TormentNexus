package tools

import (
	"context"
	"fmt"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	if sql == "" {
		return err("sql is required")
}

	result := fmt.Sprintf("Query executed: %s\nRows: id=1, name=test", sql)
	return success(result)
}

func HandleExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	if sql == "" {
		return err("sql is required")
}

	result := fmt.Sprintf("Execute executed: %s\nAffected rows: 1", sql)
	return success(result)
}