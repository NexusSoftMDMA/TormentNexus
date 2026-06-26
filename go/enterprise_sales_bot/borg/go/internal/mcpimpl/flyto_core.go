package mcpimpl

import (
	"context"
)

func HandleGetFlightInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	flight, _ :=getString(args, "flight_number")
	if flight == "" {
		return err("flight_number is required")
}

	return success("Flight " + flight + " is on time")
}

func HandleGetAirportInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "airport_code")
	if code == "" {
		return err("airport_code is required")
}

	return ok("Airport " + code + ": gates open, weather clear")
}