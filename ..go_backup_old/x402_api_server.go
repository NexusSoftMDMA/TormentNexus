package tools

import (
	"context"
)

// HandleX processes X402 API requests.
func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	return success(prompt)
}