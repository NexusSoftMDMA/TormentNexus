package tools

import (
	"context"
)

func HandleConvert(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	query := "{ " + text + " }"
	return success(query)
}

func HandleSchema(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	schema := "type Query { " + text + ": String }"
	return success(schema)
}