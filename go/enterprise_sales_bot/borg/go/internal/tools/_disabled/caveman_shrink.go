package tools

import (
	"context"
	"strings"
)

func HandleShrink(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("No text provided")
}

	words := strings.Fields(text)
	var compressed strings.Builder
	for i, w := range words {
		if i > 0 {
			compressed.WriteByte(' ')

		if len(w) > 0 {
			compressed.WriteByte(w[0])

	}
	return ok(compressed.String())
}
}
}