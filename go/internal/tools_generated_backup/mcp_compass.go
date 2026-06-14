package tools

import (
	"context"
)

func HandleGetCompassDirection(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	degrees, _ :=getInt(args, "degrees")
	degrees = degrees % 360
	if degrees < 0 {
		degrees += 360
	}
	directions := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	index := ((degrees*100)+1125)/2250%16
	return ok(directions[index])
}