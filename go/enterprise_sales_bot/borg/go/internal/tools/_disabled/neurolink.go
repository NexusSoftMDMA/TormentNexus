package tools

import (
	"context"
	"fmt"
)

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		msg = "pong"
	}
	return ok(fmt.Sprintf("Pong: %s", msg))
}