package mcpimpl

import "context"

func HandleGetVersion_mcphub_nvim(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Mcphub.Nvim v1.0.0")
}

func HandlePing_mcphub_nvim(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}