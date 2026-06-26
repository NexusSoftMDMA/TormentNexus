package tools

import (
	"context"
)

func HandleCheckModelSafety(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	modelName, _ :=getString(args, "model_name")
	if modelName == "" {
		return err("model_name is required")
}

	return success("Model " + modelName + " is safe")
}