package tools

import (
	"context"
	"fmt"
	"strings"
)

func HandleCountTokens(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	tokens := len(strings.Fields(text))
	return success(fmt.Sprintf("Estimated token count: %d", tokens))
}