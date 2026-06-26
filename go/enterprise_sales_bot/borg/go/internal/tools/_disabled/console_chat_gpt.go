package tools

import "context"

func HandleChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message parameter is required")
}

	return ok("You said: " + message)
}