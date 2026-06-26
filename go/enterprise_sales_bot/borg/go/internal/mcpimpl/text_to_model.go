package mcpimpl

import (
	"context"
)

func HandleTextToModel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text argument is required")
}

	modelType, _ :=getString(args, "model_type")
	if modelType == "" {
		modelType = "default"
	}
	return success("Converted text '" + text + "' to model type " + modelType)
}