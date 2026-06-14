package tools

import (
	"context"
)

func HandleMemoryDetective(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return success("Memory detective result for query: " + query)
}