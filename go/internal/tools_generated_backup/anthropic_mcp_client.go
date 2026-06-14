package tools

import "context"

func HandleHold(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	reason, _ :=getString(args, "reason")
	if reason == "" {
		return err("reason is required")
}

	return success("security hold placed: " + reason)
}

func HandleRelease(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	reason, _ :=getString(args, "reason")
	if reason == "" {
		return err("reason is required")
}

	return success("security hold released: " + reason)
}