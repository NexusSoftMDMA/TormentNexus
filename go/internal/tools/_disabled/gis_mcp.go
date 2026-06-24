package tools

import (
	"context"
	"strconv"
)

func HandleGis(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	latStr, _ :=getString(args, "latitude")
	lonStr, _ :=getString(args, "longitude")

	lat, e := strconv.ParseFloat(latStr, 64)
	if e != nil {
		return err("invalid latitude")
	}
	lon, e := strconv.ParseFloat(lonStr, 64)
	if e != nil {
		return err("invalid longitude")
	}

	return ok("coordinates: " + strconv.FormatFloat(lat, 'f', 6, 64) + ", " + strconv.FormatFloat(lon, 'f', 6, 64))
}