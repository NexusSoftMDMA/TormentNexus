//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
)

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tables := []string{"example_table"}
	return ok(fmt.Sprintf("Tables: %v", tables))
}

func HandleDescribeTable(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tableName, _ :=getString(args, "tableName")
	if tableName == "" {
		return err("tableName is required")
}

	return ok(fmt.Sprintf("Table %s: schema dummy", tableName))
}
