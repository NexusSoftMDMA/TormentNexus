package tools

import (
	"context"
	"fmt"
	"strings"
)

func HandleWordCount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	words := len(strings.Fields(text))
	lines := 1
	if text == "" {
		lines = 0
	} else {
		lines = strings.Count(text, "\n") + 1
	}
	chars := len(text)
	result := fmt.Sprintf("Words: %d\nLines: %d\nCharacters: %d", words, lines, chars)
	return success(result)
}