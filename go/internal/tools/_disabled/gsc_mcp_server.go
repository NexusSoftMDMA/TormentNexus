package tools

import (
	"context"
	"fmt"
)

func HandleSearchAnalytics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	site, _ :=getString(args, "site")
	query, _ :=getString(args, "query")
	return ok(fmt.Sprintf("Search analytics for %s with query %q returned 1000 clicks", site, query))
}

func HandleURLInspection(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	return success(fmt.Sprintf("URL inspection for %s: status OK, indexed", url))
}