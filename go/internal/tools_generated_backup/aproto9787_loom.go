package tools

import "context"

func HandleLoomStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	mode, _ :=getString(args, "mode")
	return success("Loom agent control plane is active. Mode: " + mode)
}