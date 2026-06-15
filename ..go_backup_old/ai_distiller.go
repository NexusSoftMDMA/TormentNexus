package tools

import (
	"context"
)

func HandleDistill(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text input required")
}

	maxLen, _ :=getInt(args, "max_length")
	if maxLen <= 0 {
		maxLen = 200
	}
	if len(text) > maxLen {
		text = text[:maxLen] + "..."
	}
	return ok(text)
}