package tools

import (
	"context"
	"strings"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return ok(message)
}

func HandleReverse(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	var builder strings.Builder
	for i := len(text) - 1; i >= 0; i-- {
		builder.WriteByte(text[i])

	return success(builder.String())
}
}