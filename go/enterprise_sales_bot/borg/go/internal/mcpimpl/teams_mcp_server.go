package mcpimpl

import (
	"context"
)

func HandleSendDirectMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	recipient, _ :=getString(args, "recipient")
	message, _ :=getString(args, "message")
	if recipient == "" {
		return err("recipient is required")
}

	if message == "" {
		return err("message is required")
}

	return ok("Direct message sent to " + recipient)
}

func HandleListDirectMessages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("[]")
}