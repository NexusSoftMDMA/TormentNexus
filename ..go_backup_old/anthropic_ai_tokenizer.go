package tools

import (
	"context"
	"fmt"
)

func HandleTokenize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	tokenCount := len(text) / 4
	return ok(fmt.Sprintf("Token count: %d", tokenCount))
}