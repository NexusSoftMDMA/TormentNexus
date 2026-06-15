//go:build ignore
// +build ignore

package tools

import (
	"context"
)

func HandleGetPrompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {

		return success(map[string]interface{}{
	}),
}
}