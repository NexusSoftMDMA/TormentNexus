package mcpimpl

import "context"

func HandleEcho_nash_arena(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}

func HandlePing_nash_arena(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong")
}