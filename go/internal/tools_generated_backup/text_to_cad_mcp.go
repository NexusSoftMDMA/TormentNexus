package tools

import (
	"context"
	"fmt"
)

func HandleTextToCad(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	cad := fmt.Sprintf("Generated CAD for: %s\n(Simulated layer output)", text)
	return ok(cad)
}