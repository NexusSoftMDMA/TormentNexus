package mcpimpl

import "context"

func HandleDebug_claude_debugs_for_you(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code argument is required")
}

	return ok("Debug output: " + code)
}