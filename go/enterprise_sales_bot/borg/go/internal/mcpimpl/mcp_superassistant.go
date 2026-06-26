package mcpimpl

import (
	"context"
	"fmt"
	"time"
)

func HandleGreet_mcp_superassistant(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok(fmt.Sprintf("Hello, %s!", name))
}

func HandleTime_mcp_superassistant(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok(time.Now().Format(time.RFC3339))
}