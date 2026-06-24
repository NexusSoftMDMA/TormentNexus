package tools

import (
	"context"
	"fmt"
)

func HandleSearchPlaces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	result := fmt.Sprintf("Foursquare places for: %s", query)
	return ok(result)
}