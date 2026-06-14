package tools

import "context"

func HandleDeepchat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return ok("You said: " + msg)
}