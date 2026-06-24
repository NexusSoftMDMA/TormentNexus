package tools

import "context"

func HandleExecuteAction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	return ok("Executed: " + prompt)
}