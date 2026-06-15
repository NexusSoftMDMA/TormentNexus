//go:build ignore
// +build ignore

package tools

import (
	"context"
)

func HandleTouchGrass(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("You touched the grass. 🌿 Your soul is replenished.")
}
