package mcpimpl

import (
	"context"
	"fmt"
)

func HandleSearchNutrient(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	result := fmt.Sprintf(`{"food":"%s","calories":100,"protein":"10g"}`, query)
	return success(result)
}