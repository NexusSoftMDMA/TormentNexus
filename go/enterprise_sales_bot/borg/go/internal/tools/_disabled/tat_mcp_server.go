package tools

import (
	"context"
	"time"
)

func HandleGetTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	format, _ :=getString(args, "format")
	now := time.Now()
	var result string
	if format == "" {
		result = now.Format("2006-01-02 15:04:05")
	} else {
		result = now.Format(format)

	return ok(result)
}
}