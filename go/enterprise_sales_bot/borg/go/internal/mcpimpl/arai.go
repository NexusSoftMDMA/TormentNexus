package mcpimpl

import (
	"context"
	"fmt"
)

func HandleArai(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok(fmt.Sprintf("Hello, %s! from Arai", name))
}