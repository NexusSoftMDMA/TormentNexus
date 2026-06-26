package tools

import (
	"context"
	"fmt"
)

func HandleProbe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	return ok(fmt.Sprintf("Probed: %s", prompt))
}