package tools

import (
	"context"
	"time"
)

func HandleGetCurrentTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	format, _ :=getString(args, "format")
	if format == "" {
		format = time.RFC3339
	}
	now := time.Now().Format(format)
	return ok(now)
}

func HandleFormatUnixTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	timestamp, _ :=getInt(args, "timestamp")
	if timestamp == 0 {
		return err("timestamp is required")
}

	t := time.Unix(int64(timestamp), 0)
	format, _ :=getString(args, "format")
	if format == "" {
		format = time.RFC3339
	}
	return ok(t.Format(format))
}