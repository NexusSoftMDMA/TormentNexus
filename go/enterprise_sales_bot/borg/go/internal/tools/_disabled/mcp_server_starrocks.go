package tools

import (
	"context"
	"fmt"
)

func HandleExecuteSQL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	return success(fmt.Sprintf("Executed SQL: %s", sql))
}