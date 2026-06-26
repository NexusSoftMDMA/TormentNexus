package mcpimpl

import (
	"context"
	"time"
)

func HandleGetCurrentTime_mcp_time(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().UTC()
	return ok(now.Format(time.RFC3339))
}

func HandleGetTimeInTimezone(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tz, _ :=getString(args, "timezone")
	if tz == "" {
		return err("timezone is required")
}

	loc, e := time.LoadLocation(tz)
	if e != nil {
		return err("invalid timezone: " + e.Error())
}

	now := time.Now().In(loc)
	return ok(now.Format(time.RFC3339))
}