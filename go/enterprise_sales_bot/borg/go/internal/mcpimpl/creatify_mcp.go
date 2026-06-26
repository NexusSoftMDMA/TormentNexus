package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGenerateText_creatify_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	return success(fmt.Sprintf("Generated text based on: %s", prompt))
}

func HandleGenerateImage_creatify_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	return success(fmt.Sprintf("Generated image based on: %s", prompt))
}