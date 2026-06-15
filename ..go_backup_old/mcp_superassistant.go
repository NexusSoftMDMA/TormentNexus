package tools

import (
	"context"
	"fmt"
	"time"
)

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok(fmt.Sprintf("Hello, %s!", name))
}

func HandleTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok(time.Now().Format(time.RFC3339))
}