package tools

import (
	"context"
	"fmt"
)

func HandleClick(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	x, _ :=getInt(args, "x")
	y, _ :=getInt(args, "y")
	return ok(fmt.Sprintf("clicked at (%d,%d)", x, y))
}

func HandleType(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	return ok("typed \"" + text + "\"")
}