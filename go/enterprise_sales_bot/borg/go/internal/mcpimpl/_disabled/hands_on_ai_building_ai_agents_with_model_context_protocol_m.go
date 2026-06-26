package mcpimpl

import (
	"context"
	"fmt"
	"time"
)

func HandleCurrentTime_hands_on_ai_building_ai_agents_with_model_context_protocol_m(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return ok("Current time: " + now)
}

func HandleEcho_hands_on_ai_building_ai_agents_with_model_context_protocol_m(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("Missing 'message' argument")
}

	return ok("Echo: " + msg)
}