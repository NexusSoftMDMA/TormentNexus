package mcpimpl

import "context"

func HandleMcporterStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("mcporter server is running")
}

func HandleMcporterEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return ok(msg)
}