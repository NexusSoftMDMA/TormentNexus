package mcpimpl

import "context"

func HandleChat_mcp_server_qwen_max(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return success(msg)
}