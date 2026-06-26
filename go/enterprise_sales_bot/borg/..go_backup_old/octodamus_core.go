package tools

import (
	"context"
	"time"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok("echo: " + msg)
}

func HandleTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	return success("Current time: " + now)
}