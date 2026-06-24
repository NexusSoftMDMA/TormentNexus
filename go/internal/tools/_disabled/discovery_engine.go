package tools

import (
	"context"
	"fmt"
)

func HandleList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	return ok(fmt.Sprintf("listing items, limit=%d", limit))
}

func HandleGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	return ok("getting item: " + id)
}