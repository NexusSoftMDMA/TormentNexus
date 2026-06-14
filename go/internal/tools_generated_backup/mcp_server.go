package tools

import (
	"context"
	"time"
)

func HandleHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello, " + name + "!")
}

func HandleTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format(time.RFC1123)
	return ok("Current time: " + now)
}