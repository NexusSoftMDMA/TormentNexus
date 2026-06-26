package tools

import (
	"context"
	"fmt"
)

func HandleGenerateImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	return success(fmt.Sprintf("Generated image for prompt: %s", prompt))
}