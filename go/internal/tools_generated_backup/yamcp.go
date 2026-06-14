package tools

import (
	"context"
	"time"
)

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return success("Hello, " + name + "!")
}

func HandleTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return success("Current time: " + now)
}