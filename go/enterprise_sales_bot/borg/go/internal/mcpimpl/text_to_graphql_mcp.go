package mcpimpl

import (
	"context"
)

func HandleConvert_text_to_graphql_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	query := "{ " + text + " }"
	return success(query)
}

func HandleSchema(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	schema := "type Query { " + text + ": String }"
	return success(schema)
}