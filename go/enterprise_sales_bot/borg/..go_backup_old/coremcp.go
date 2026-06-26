package tools

import (
	"context"
	"time"
)

// HandleGetTime returns the current time in RFC3339 format.
func HandleGetTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok(time.Now().Format(time.RFC3339))
}

// HandleEcho returns the provided message argument.
func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}