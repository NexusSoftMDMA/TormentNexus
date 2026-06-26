package mcpimpl

import (
	"context"
	"fmt"
)

func HandleKom(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok(fmt.Sprintf("Hello, %s!", name))
}