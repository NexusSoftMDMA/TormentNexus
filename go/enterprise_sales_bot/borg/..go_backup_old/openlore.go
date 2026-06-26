package tools

import (
	"context"
	"fmt"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	msg := fmt.Sprintf("Openlore knowledge graph query: %s", query)
	return ok(msg)
}