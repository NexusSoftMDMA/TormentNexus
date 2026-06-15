package tools

import (
	"context"
	"fmt"
	"os"
)

func HandleGetContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value := os.Getenv(key)
	if value == "" {
		return ok("Environment variable '" + key + "' is not set")
}

	return ok(fmt.Sprintf("Environment variable '%s' = '%s'", key, value))
}