//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return success("Query received: " + query)
}

func HandleCount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	count, _ :=getInt(args, "count")
	return success("Count is: " + fmt.Sprintf("%d", count))
}
