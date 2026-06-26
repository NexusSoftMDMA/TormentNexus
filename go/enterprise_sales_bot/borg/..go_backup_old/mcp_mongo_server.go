package tools

import (
	"context"
)

func HandleListDatabases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dbName, _ :=getString(args, "database")
	if dbName == "" {
		return err("database parameter required")
}

	return ok("Listed databases for: " + dbName)
}

func HandleFindDocuments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collection, _ :=getString(args, "collection")
	if collection == "" {
		return err("collection parameter required")
}

	return success("Found documents in " + collection)
}