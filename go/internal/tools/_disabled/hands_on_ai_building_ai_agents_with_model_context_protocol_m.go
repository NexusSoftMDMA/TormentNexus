package tools

import (
	"context"
	"fmt"
	"time"
)

func HandleCurrentTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return ok("Current time: " + now)
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("Missing 'message' argument")
}

	return ok("Echo: " + msg)
}