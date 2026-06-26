package mcpimpl

import "context"

func HandleExecuteQuery_berthojoris_mcp_mysql_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	project, _ :=getString(args, "project")
	if query == "" || project == "" {
		return err("query and project are required")
}

	return ok("query executed for project " + project)
}

func HandleSelectQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	project, _ :=getString(args, "project")
	if query == "" || project == "" {
		return err("query and project are required")
}

	return success("select query executed, results: (mock data)")
}