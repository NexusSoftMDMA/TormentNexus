package tools

import (
	"context"
)

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong")
}

func HandlePong(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("ping")
}