package mcpimpl

import (
	"context"
	"fmt"
)

func HandleQuery_mongodb_lens(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	database, _ :=getString(args, "database")
	collection, _ :=getString(args, "collection")
	filter, _ :=getString(args, "filter")
	result := fmt.Sprintf("Queried %s.%s with filter %s", database, collection, filter)
	return ok(result)
}

func HandleListCollections_mongodb_lens(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	database, _ :=getString(args, "database")
	result := fmt.Sprintf("Listed collections in %s", database)
	return ok(result)
}