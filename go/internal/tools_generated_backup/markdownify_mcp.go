package tools

import (
	"context"
)

func HandleConvertToMarkdown(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "content")
	if text == "" {
		return err("content is required")
	}
	markdown := "```\n" + text + "\n```"
	return success(markdown)
}