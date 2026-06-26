package tools

import (
	"context"
)

func HandleA2cr(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	result := input + "Cr"
	return success(result)
}