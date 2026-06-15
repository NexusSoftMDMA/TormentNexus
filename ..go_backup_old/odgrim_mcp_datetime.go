package tools

import (
    "context"
    "time"
)

func HandleGetCurrentDateTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    format, _ :=getString(args, "format")
    if format == "" {
        format = time.RFC3339
    }
    return ok(time.Now().Format(format))
}