package mcpimpl

import (
	"context"
	"fmt"
)

func HandleSearchRestaurants(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("Missing query parameter")
}

	return success(fmt.Sprintf("Search results for '%s': [Restaurant A, Restaurant B]", query))
}

func HandleGetMenu(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "restaurant_id")
	if id == 0 {
		return err("Missing or invalid restaurant_id")
}

	return success(fmt.Sprintf("Menu for restaurant %d: [Item1: $5, Item2: $10]", id))
}