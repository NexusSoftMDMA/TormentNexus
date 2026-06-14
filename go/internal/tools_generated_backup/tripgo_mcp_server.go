package tools

import (
	"context"
	"fmt"
)

func HandleGetTrip(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	origin, _ :=getString(args, "origin")
	dest, _ :=getString(args, "destination")
	msg := fmt.Sprintf("Trip from %s to %s found.", origin, dest)
	return ok(msg)
}

func HandleGetStopInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	stopID, _ :=getString(args, "stop_id")
	msg := fmt.Sprintf("Stop info for %s retrieved.", stopID)
	return ok(msg)
}