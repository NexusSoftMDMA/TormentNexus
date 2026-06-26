package tools

import (
    "context"
    "fmt"
)

func HandlePcq(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    if query == "" {
        return err("query is required")
}

    limit, _ :=getInt(args, "limit")
    if limit > 0 {
        return success(fmt.Sprintf("Query: %s (limit %d)", query, limit))
}

    return success(fmt.Sprintf("Query: %s", query))
}