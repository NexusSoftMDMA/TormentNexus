package tools

import (
	"context"
)

func HandleCapture(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Video capture initiated")
}