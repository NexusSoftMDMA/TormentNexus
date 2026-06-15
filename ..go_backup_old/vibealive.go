package tools

import (
	"context"
)

func HandleVibeCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	mood, _ :=getString(args, "mood")
	if mood == "" {
		mood = "Neutral"
	}
	return ok("Vibealive: mood is " + mood)
}