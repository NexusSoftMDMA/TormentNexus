package mcpimpl

import (
	"context"
	"fmt"
	"time"
)

func HandleGetCurrentTime_guanxiong_mcp_server_time(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tz, _ :=getString(args, "timezone")
	if tz == "" {
		tz = "UTC"
	}
	loc, e := time.LoadLocation(tz)
	if e != nil {
		return err("Invalid timezone: " + tz)
}

	now := time.Now().In(loc)
	return ok(now.Format("2006-01-02 15:04:05 MST"))
}

func HandleGetSupportedTimezones(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Use IANA timezone names like America/New_York, Asia/Shanghai, etc.")
}