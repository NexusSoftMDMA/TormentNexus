package mcpimpl

import "context"

func HandleStartSession(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	getInt(args, "duration") // ignored, default 25
	return ok("Pomodoro session started")
}

func HandleStopSession(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Pomodoro session stopped")
}