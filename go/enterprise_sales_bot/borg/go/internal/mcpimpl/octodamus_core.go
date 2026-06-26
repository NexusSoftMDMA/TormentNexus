package mcpimpl

import (
	"context"
	"time"
)

func HandleEcho_octodamus_core(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok("echo: " + msg)
}

func HandleTime_octodamus_core(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	return success("Current time: " + now)
}