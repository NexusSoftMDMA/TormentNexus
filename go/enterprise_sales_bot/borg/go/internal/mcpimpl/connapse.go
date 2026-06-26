package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGreet_connapse(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok(fmt.Sprintf("Hello, %s!", name))
}