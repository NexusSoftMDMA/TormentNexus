package tools

import "context"

func HandleDandan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return success("Dandan says: " + message)
}