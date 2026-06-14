package tools

import (
	"context"
	"time"
)

func HandleGetCurrentTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tz, _ :=getString(args, "timezone")
	if tz == "" {
		tz = "UTC"
	}
	loc, e := time.LoadLocation(tz)
	if e != nil {
		return err("invalid timezone: " + e.Error())
}

	now := time.Now().In(loc)
	return success("Current time: " + now.Format(time.RFC3339))
}