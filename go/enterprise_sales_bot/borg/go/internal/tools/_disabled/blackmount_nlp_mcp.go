package tools

import (
	"context"
	"fmt"
	"strings"
)

func HandleWordCount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	words := strings.Fields(text)
	return ok(fmt.Sprintf("Word count: %d", len(words)))
}

func HandleKeywordSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	keyword, _ :=getString(args, "keyword")
	if text == "" || keyword == "" {
		return err("text and keyword are required")
}

	found := strings.Contains(strings.ToLower(text), strings.ToLower(keyword))
	if found {
		return ok("Keyword found")
}

	return ok("Keyword not found")
}