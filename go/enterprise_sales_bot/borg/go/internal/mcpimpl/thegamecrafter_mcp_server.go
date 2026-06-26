package mcpimpl

import (
	"context"
	"fmt"
)

func HandleSearchGames_thegamecrafter_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok(fmt.Sprintf("Searched for games: %s", query))
}

func HandleGetGameDetails_thegamecrafter_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	return ok(fmt.Sprintf("Game details for ID: %s", id))
}