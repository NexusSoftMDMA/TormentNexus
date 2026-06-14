package tools

import "context"

func HandleLookAt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat, _ :=getString(args, "latitude")
	lng, _ :=getString(args, "longitude")
	height, _ :=getString(args, "height")
	if lat == "" || lng == "" {
		return err("latitude and longitude required")
}

	return ok("Camera moved to " + lat + ", " + lng + " at height " + height)
}

func HandleFlyTo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	entityId, _ :=getString(args, "entityId")
	duration, _ :=getString(args, "duration")
	if entityId == "" {
		return err("entityId required")
}

	return ok("Flying to " + entityId + " with duration " + duration)
}