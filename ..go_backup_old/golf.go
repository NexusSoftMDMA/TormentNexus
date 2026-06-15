package tools

import (
	"context"
	"fmt"
)

func HandleGetScore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	player, _ :=getString(args, "player")
	return ok(fmt.Sprintf("Score for %s is 72", player))
}

func HandlePostScore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	player, _ :=getString(args, "player")
	score, _ :=getInt(args, "score")
	return success(fmt.Sprintf("Recorded score %d for %s", score, player))
}