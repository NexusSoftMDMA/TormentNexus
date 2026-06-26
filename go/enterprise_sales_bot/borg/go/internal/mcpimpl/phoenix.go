package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGreeting_phoenix(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return success(fmt.Sprintf("Hello, %s! Welcome to Phoenix.", name))
}