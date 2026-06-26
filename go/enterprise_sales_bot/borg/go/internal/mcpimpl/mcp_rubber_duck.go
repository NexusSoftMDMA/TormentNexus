package mcpimpl

import "context"

func HandleTalkToDuck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return ok("🦆 Quack! You said: " + message)
}