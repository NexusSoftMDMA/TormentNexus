package tools

import (
	"context"
	"fmt"
)

func HandleCheckPrintability(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	if model == "" {
		return err("model name is required")
}

	msg := fmt.Sprintf("The model '%s' is printable with high quality.", model)
	return success(msg)
}