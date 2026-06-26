package tools

import "context"

func HandleGetVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Mcphub.Nvim v1.0.0")
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}