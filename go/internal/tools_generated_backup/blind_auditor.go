package tools

import "context"

func HandleAudit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("no text provided")
}

	return success("Audit completed for: " + text)
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}