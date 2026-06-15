package tools

import (
	"context"
	"strconv"
	"strings"
)

func HandleCountTokens(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	tokens := strings.Fields(text)
	count := len(tokens)
	return success("Token count: " + strconv.Itoa(count))
}