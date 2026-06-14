package tools

import (
	"context"
	"time"
)

func HandleCreateEvent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	summary, _ :=getString(args, "summary")
	start, _ :=getString(args, "startDateTime")
	end, _ :=getString(args, "endDateTime")
	if summary == "" || start == "" || end == "" {
		return err("missing required fields: summary, startDateTime, endDateTime")
}

	return success("Event created: " + summary + " from " + start + " to " + end)
}

func HandleCalculateDuration(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	startStr, _ :=getString(args, "start")
	endStr, _ :=getString(args, "end")
	if startStr == "" || endStr == "" {
		return err("missing start or end time")
}

	startTime, e := time.Parse(time.RFC3339, startStr)
	if e != nil {
		return err("invalid start time format: " + e.Error())
}

	endTime, e := time.Parse(time.RFC3339, endStr)
	if e != nil {
		return err("invalid end time format: " + e.Error())
}

	duration := endTime.Sub(startTime)
	return ok("Duration: " + duration.String())
}