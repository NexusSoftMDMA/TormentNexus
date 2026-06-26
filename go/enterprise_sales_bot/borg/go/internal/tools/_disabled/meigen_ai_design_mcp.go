package tools

import (
	"context"
	"fmt"
)

func HandleQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("The only way to do great work is to love what you do. - Steve Jobs")
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return ok(fmt.Sprintf("You said: %s", msg))
}