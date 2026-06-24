package tools

import (
    "context"
    "fmt"
)

func HandleQueryGuide(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    if query == "" {
        return err("query is required")
}

    return success(fmt.Sprintf("PostgreSQL guide for '%s': Use proper indexing, avoid N+1 queries, prefer CTEs for complex joins.", query))
}