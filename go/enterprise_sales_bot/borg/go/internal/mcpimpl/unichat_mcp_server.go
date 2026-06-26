package mcpimpl

import "context"

func HandleListModels_unichat_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success(`["gpt-3.5","gpt-4"]`)
}

func HandleSendChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return ok("You said: " + message)
}