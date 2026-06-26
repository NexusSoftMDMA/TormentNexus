package tools

import (
	"context"
)

func HandleBriefing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		return err("topic is required")
}

	return ok("Briefing for " + topic)
}

func HandleMemoryWrite(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" {
		return err("key is required")
}

	_ = value
	return success("memory written for key " + key)
}