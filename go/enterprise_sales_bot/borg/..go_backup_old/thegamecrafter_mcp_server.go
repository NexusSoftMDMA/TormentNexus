package tools

import (
	"context"
	"fmt"
)

func HandleSearchGames(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok(fmt.Sprintf("Searched for games: %s", query))
}

func HandleGetGameDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	return ok(fmt.Sprintf("Game details for ID: %s", id))
}