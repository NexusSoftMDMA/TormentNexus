package tools

import (
	"context"
)

func HandleGetLastMinuteDeals(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	if category == "" {
		return err("category is required")
}

	return success(`{"deals": [{"title": "Last minute deal in ` + category + `", "discount": 30}]}`)
}

func HandleSearchDeals(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return ok("Search results for: " + query)
}