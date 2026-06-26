package tools

import (
	"context"
	"fmt"
)

func HandleSearchInvestments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	result := fmt.Sprintf("Search results for %s: ...", query)
	return ok(result)
}// touch 1781132130
