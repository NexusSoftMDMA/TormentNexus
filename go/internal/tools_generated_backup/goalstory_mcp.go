package tools

import (
	"context"
)

func HandleGetGoal(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	return success("Goal " + id + " retrieved")
}

func HandleListGoals(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Goals list")
}