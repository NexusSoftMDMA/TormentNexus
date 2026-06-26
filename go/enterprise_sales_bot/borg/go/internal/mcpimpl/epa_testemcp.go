package mcpimpl

import (
	"context"
)

func HandleSearch_epa_testemcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return success("EPA search result for: " + query)
}