package tools

import (
	"context"
	"fmt"
	"time"
)

func HandleGetCurrentTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	locName, _ :=getString(args, "timezone")
	loc := time.UTC
	if locName != "" {
		var e error
		loc, e = time.LoadLocation(locName)
		if e != nil {
			return err("invalid timezone: " + e.Error())

	}
	now := time.Now().In(loc)
	return ok(now.Format(time.RFC3339))
}

}

func HandleConvertTimezone(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	timeStr, _ :=getString(args, "time")
	fromZone, _ :=getString(args, "from_timezone")
	toZone, _ :=getString(args, "to_timezone")
	if timeStr == "" || fromZone == "" || toZone == "" {
		return err("missing required parameters: time, from_timezone, to_timezone")
}

	fromLoc, e := time.LoadLocation(fromZone)
	if e != nil {
		return err("invalid from_timezone: " + e.Error())
}

	toLoc, e := time.LoadLocation(toZone)
	if e != nil {
		return err("invalid to_timezone: " + e.Error())
}

	t, e := time.ParseInLocation(time.RFC3339, timeStr, fromLoc)
	if e != nil {
		return err("invalid time format, expected RFC3339: " + e.Error())
}

	converted := t.In(toLoc)
	return ok(converted.Format(time.RFC3339))
}