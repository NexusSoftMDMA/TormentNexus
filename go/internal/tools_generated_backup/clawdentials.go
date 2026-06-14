package tools

import (
	"context"
)

func HandleGetCredential(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	// Simulate credential retrieval
	return ok("credential for " + key + " is secret123")
}