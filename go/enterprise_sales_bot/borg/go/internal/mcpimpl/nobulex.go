package mcpimpl

import "context"

func HandleEcho_nobulex(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return ok(msg)
}

func HandleVersion_nobulex(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Nobulex v1.0.0")
}