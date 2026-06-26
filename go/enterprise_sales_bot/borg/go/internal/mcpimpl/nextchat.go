package mcpimpl

import (
	"context"
	"fmt"
)

func HandleListModels_nextchat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok(`{"models": ["gpt-3.5", "gpt-4"]}`)
}

func HandleChat_nextchat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	return success(fmt.Sprintf("Echo: %s", prompt))
}