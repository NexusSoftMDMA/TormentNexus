package tools

import "context"

func HandleGetContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Context provided successfully")
}