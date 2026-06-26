package mcpimpl

import (
	"context"
	"fmt"
)

func HandleLookupFunction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok(fmt.Sprintf("Function '%s' found", name))
}