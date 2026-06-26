package tools

import (
    "context"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    if query == "" {
        return err("query is required")
}

    return ok("Searching Vertex AI for: " + query)
}