package mcpimpl

import (
    "context"
)

func HandleListTables_sqlite_explorer_fastmcp_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("Tables: users, posts")
}

func HandleRunQuery_sqlite_explorer_fastmcp_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "sql")
    if query == "" {
        return err("sql parameter is required")
}

    return success("Query executed: " + query)
}