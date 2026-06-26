package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGenerateImage_imagen3_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	return success(fmt.Sprintf("Image generation started for: %s", prompt))
}