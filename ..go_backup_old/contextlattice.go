package tools

import (
	"context"
	"time"
)

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok("Hello, " + name + "!")
}

func HandleTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Current time: " + time.Now().Format(time.RFC3339))
}