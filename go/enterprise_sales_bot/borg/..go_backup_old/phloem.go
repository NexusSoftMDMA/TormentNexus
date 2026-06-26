package tools

import "context"

func HandlePhloem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message parameter is required")
}

	return ok("You said: " + msg)
}