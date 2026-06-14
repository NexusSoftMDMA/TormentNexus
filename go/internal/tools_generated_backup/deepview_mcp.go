package tools

import "context"

func HandleDeepview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("input is required")
}

	message := "Deepview analysis for: " + input
	return ok(message)
}