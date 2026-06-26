package mcpimpl

import (
	"context"
)

func HandleSelvageInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return success("Selvage info for: " + query)
}