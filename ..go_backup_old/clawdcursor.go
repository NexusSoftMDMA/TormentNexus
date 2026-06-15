package tools

import (
	"context"
	"fmt"
)

func HandleClick(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	x, _ :=getInt(args, "x")
	y, _ :=getInt(args, "y")
	return ok(fmt.Sprintf("Clicked at (%d, %d)", x, y))
}

func HandleType(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	return ok(fmt.Sprintf("Typed text: %s", text))
}