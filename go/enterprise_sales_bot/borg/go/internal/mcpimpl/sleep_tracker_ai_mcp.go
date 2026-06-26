package mcpimpl

import (
	"context"
	"fmt"
)

func HandleLogSleep(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	duration, _ :=getInt(args, "duration")
	quality, _ :=getInt(args, "quality")
	date, _ :=getString(args, "date")
	return ok(fmt.Sprintf("Logged sleep: %d min, quality %d/10, date %s", duration, quality, date))
}

func HandleGetSleepStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Sleep stats: avg 7.5h, avg quality 8/10")
}