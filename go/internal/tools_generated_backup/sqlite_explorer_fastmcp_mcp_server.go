package tools

import (
    "context"
)

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("Tables: users, posts")
}

func HandleRunQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "sql")
    if query == "" {
        return err("sql parameter is required")
}

    return success("Query executed: " + query)
}