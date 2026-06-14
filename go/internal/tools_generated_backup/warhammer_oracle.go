package tools

import (
	"context"
)

func HandleWarhammerQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("Warhammer Oracle requires a query")
}

	return ok("Warhammer Oracle response: " + query)
}