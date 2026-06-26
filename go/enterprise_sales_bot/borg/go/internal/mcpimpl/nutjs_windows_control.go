package mcpimpl

import (
	"context"
	"fmt"
)

func HandleMouseMove(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	x, _ :=getInt(args, "x")
	y, _ :=getInt(args, "y")
	return ok(fmt.Sprintf("Mouse moved to (%d, %d)", x, y))
}

func HandleClick_nutjs_windows_control(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	button, _ :=getString(args, "button")
	if button == "" {
		button = "left"
	}
	return ok(fmt.Sprintf("Clicked %s button", button))
}