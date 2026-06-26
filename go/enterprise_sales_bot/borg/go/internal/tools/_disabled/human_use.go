package tools

import "context"

func HandleHumanUse(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return success(message)
}