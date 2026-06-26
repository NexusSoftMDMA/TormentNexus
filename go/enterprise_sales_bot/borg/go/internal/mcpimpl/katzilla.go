package mcpimpl

import (
	"context"
	"fmt"
	"time"
)

func HandleGreet_katzilla(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return success(fmt.Sprintf("Hello, %s!", name))
}

func HandleTime_katzilla(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return ok(fmt.Sprintf("Current time is %s", now))
}