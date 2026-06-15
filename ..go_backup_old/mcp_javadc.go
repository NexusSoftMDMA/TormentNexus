package tools

import (
	"context"
)

func HandleJavadocLookup(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return success("Javadoc query: " + query)
}