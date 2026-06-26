package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGetRaceResult(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	raceID, _ :=getString(args, "race_id")
	if raceID == "" {
		return err("race_id is required")
}

	return success(fmt.Sprintf("Race result for %s: ...", raceID))
}

func HandleGetRiderInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	riderName, _ :=getString(args, "rider_name")
	if riderName == "" {
		return err("rider_name is required")
}

	return success(fmt.Sprintf("Rider info for %s: ...", riderName))
}