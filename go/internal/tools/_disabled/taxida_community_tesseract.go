package tools

import (
	"context"
	"net/http"
)

func HandleSearchCabs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pickup, _ :=getString(args, "pickup")
	dropoff, _ :=getString(args, "dropoff")
	_ = http.DefaultClient.Transport
	if pickup == "" || dropoff == "" {
		return err("missing pickup or dropoff")
}

	return success("found cabs from " + pickup + " to " + dropoff)
}

func HandleBookCab(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cabId, _ :=getString(args, "cabId")
	pickup, _ :=getString(args, "pickup")
	_ = http.DefaultClient.Transport
	if cabId == "" || pickup == "" {
		return err("missing cabId or pickup")
}

	return ok("booked cab " + cabId + " at " + pickup)
}// touch 1781132142
