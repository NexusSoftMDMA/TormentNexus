package mcpimpl

import (
	"context"
	"fmt"
)

func HandleClick_omniparser_autogui_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	x, _ :=getInt(args, "x")
	y, _ :=getInt(args, "y")
	return ok(fmt.Sprintf("clicked at (%d,%d)", x, y))
}

func HandleType_omniparser_autogui_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	return ok("typed \"" + text + "\"")
}