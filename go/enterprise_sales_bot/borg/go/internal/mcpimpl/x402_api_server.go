package mcpimpl

import (
	"context"
)

// HandleX processes X402 API requests.
func HandleX_x402_api_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	return success(prompt)
}