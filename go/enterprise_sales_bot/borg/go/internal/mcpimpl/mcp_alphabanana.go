package mcpimpl

import (
	"context"
	"fmt"
)

func HandleBananaPing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Alphabanana pong!")
}

func HandleBananaEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	word, _ :=getString(args, "word")
	msg := fmt.Sprintf("Alphabanana echoes: %s (length %d)", word, len(word))
	return success(msg)
}