package mcpimpl

import (
	"context"
	"fmt"
	"strings"
)

func HandleTokenAnalyzer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	tokens := len(strings.Fields(text))
	return ok(fmt.Sprintf("Token weight: %d", tokens))
}