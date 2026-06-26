package tools

import "context"

func HandleExecuteJS(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	script, _ :=getString(args, "script")
	if script == "" {
		return err("script is required")
}

	return ok("Executed script: " + script)
}