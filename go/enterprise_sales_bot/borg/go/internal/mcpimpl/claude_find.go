package mcpimpl

import (
	"context"
	"fmt"
)

func HandleFind_claude_find(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return success(fmt.Sprintf("Found results for: %s", query))
}