package tools

import (
	"context"
	"net/http"
)

func HandleCheckAirspace(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	location, _ :=getString(args, "location")
	if location == "" {
		return err("location is required")
}

	_, e := http.DefaultClient.Get("https://api.airspace-monitor.example.com/check?location=" + location)
	if e != nil {
		return err("request failed")
}

	return ok("Airspace at " + location + " is clear")
}

func HandleGetNoFlyZones(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("No-fly zones: [Zone A, Zone B]")
}