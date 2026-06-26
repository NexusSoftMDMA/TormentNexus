package tools

import "context"

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collection, _ :=getString(args, "collection")
	_ = collection
	return success("Query executed")
}

func HandleInsert(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collection, _ :=getString(args, "collection")
	document, _ :=getString(args, "document")
	_ = collection
	_ = document
	return ok("Document inserted")
}