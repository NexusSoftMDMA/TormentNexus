package tools

import (
	"context"
	"fmt"
)

func HandleGetCurrents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		query = "latest"
	}
	return ok(fmt.Sprintf("Current news for '%s': Placeholder response from Currents Mcp server.", query))
}