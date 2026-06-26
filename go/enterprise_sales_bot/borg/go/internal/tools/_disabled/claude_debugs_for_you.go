package tools

import "context"

func HandleDebug(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code argument is required")
}

	return ok("Debug output: " + code)
}