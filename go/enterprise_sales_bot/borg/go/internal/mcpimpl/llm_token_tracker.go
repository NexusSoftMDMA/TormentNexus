package mcpimpl

import (
	"context"
	"fmt"
	"strings"
)

func HandleCountTokens_llm_token_tracker(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	tokens := len(strings.Fields(text))
	return success(fmt.Sprintf("Estimated token count: %d", tokens))
}