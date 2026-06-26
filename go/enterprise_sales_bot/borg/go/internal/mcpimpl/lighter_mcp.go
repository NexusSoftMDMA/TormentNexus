package mcpimpl

import (
	"context"
)

func HandleLighten(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	msg := "Lighter says: " + text
	return success(msg)
}

func HandleCheckLight(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	value, _ :=getInt(args, "value")
	if value < 100 {
		return ok("It is light")
}

	return err("It is heavy")
}