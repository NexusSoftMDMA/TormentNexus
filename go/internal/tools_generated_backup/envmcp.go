package tools

import (
	"context"
	"os"
)

func HandleGetEnv(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value := os.Getenv(key)
	return ok(value)
}// touch 1781132125
