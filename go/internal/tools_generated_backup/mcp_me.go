package tools

import (
	"context"
	"fmt"
)

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok(fmt.Sprintf("Hello, %s! This is Mcp Me.", name))
}

func HandleVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Mcp Me version 1.0.0")
}