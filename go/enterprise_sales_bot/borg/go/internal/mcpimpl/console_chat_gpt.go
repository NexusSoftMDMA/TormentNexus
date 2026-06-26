package mcpimpl

import "context"

func HandleChat_console_chat_gpt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message parameter is required")
}

	return ok("You said: " + message)
}