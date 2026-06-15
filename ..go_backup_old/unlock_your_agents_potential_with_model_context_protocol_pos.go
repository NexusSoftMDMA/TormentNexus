package tools

import (
	"context"
	"fmt"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	return ok(fmt.Sprintf("Executed query: %s (placeholder)", query))
}

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	result := `["users", "products", "orders", "reviews"]`
	return ok(result)
}