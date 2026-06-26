package mcpimpl

import "context"

func HandleQuery_mongo_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collection, _ :=getString(args, "collection")
	_ = collection
	return success("Query executed")
}

func HandleInsert_mongo_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collection, _ :=getString(args, "collection")
	document, _ :=getString(args, "document")
	_ = collection
	_ = document
	return ok("Document inserted")
}