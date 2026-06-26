package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleClaude(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("input is required")
}

	_ = http.DefaultClient
	return ok(fmt.Sprintf("Echo: %s", input))
}