package mcpimpl

import (
	"context"
	"fmt"
)

func HandleProbe_llmprobe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	return ok(fmt.Sprintf("Probed: %s", prompt))
}