//go:build ignore
// +build ignore

package tools

import "context"

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}

func HandleExecuteSQL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	return success("Executed: " + sql)
}
