package tools

import (
	"context"
	"fmt"
)

func HandleSearchCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return ok(fmt.Sprintf("Searching for code: %s", query))
}

func HandleGetCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	return ok(fmt.Sprintf("Getting code from: %s", path))
}