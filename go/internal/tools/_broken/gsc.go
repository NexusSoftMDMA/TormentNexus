package tools

import (
	"context"
	"os"
)

func HandleListProperties(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("GSC_ACCESS_TOKEN")
	if token == "" {
		return err("GSC_ACCESS_TOKEN env var missing")
}

	return ok("text")
}