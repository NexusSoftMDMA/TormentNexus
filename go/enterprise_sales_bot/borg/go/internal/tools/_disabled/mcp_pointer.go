package tools

import "context"

func HandleGetPointer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pointer at (100, 200)")
}