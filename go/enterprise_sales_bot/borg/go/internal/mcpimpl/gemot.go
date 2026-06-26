package mcpimpl

import (
	"context"
)

func HandleGemotInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	gemType, _ :=getString(args, "type")
	if gemType == "" {
		gemType = "unknown"
	}
	msg := "Gemot info: type = " + gemType
	return ok(msg)
}