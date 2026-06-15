package tools

import (
	"context"
	"fmt"
)

func HandleSearchDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok(fmt.Sprintf("Search results for '%s': See https://developer.garmin.com", query))
}