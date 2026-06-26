package mcpimpl

import "context"

func HandleQuery_jdbcx(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	if sql == "" {
		return err("sql is required")
}

	return ok("query executed: " + sql)
}

func HandleExecute_jdbcx(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	if sql == "" {
		return err("sql is required")
}

	return success("execute executed: " + sql)
}