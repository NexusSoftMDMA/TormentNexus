package tools

import (
	"context"
	"fmt"
)

func HandleSearchConsole(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	site, _ :=getString(args, "site")
	query, _ :=getString(args, "query")
	msg := fmt.Sprintf("Search Console query for site %s: %s", site, query)
	return ok(msg)
}