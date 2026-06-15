package tools

import (
	"context"
	"strings"
)

func HandleSift(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	items := []string{"apple", "banana", "cherry", "date", "elderberry"}
	var matches []string
	for _, item := range items {
		if strings.Contains(item, query) {
			matches = append(matches, item)

	}
	return ok(matches)
}
}