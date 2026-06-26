package tools

import "context"

func HandleAugment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("no text provided")
}

	return ok("augmented: " + text + " [MCP]")
}