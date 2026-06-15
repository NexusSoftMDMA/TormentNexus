package tools

import (
	"context"
	"fmt"
)

func HandleGoToGoal(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	x, _ :=getInt(args, "x")
	y, _ :=getInt(args, "y")
	msg := fmt.Sprintf("Agent heading straight to goal at (%d, %d)", x, y)
	return ok(msg)
}