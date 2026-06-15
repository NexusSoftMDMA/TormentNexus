package tools

import (
    "context"
    "fmt"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    msg := fmt.Sprintf("Ragmap query: %s", query)
    return ok(msg)
}

func HandleGetMap(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    id, _ :=getString(args, "id")
    msg := fmt.Sprintf("Map data for id: %s", id)
    return success(msg)
}