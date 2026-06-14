package tools

import (
	"context"
	"strings"
)

func HandleCompress(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	maxTokens, _ :=getInt(args, "maxTokens")
	if text == "" {
		return err("no text provided")
}

	if maxTokens <= 0 {
		return success(text)
}

	words := strings.Fields(text)
	if len(words) <= maxTokens {
		return success(text)
}

	compressed := strings.Join(words[:maxTokens], " ")
	return success(compressed)
}