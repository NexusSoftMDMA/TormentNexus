package tools

import (
	"context"
)

func HandleJobSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return success("Job search for: " + query)
}