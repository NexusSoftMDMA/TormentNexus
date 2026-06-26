package mcpimpl

import "context"

func HandleDicomHl7(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	if action == "" {
		return err("action is required")
}

	if action == "echo" {
		message, _ :=getString(args, "message")
		return ok("echo: " + message)
}

	return err("unknown action: " + action)
}