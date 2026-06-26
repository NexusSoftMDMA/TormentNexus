package mcpimpl

import (
	"context"
	"strings"
)

func ListHealthDataTypes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	types := []string{"stepCount", "heartRate", "activeEnergyBurned", "distanceWalkingRunning"}
	return ok("Available types: " + strings.Join(types, ", "))
}

func GetHealthData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dataType, _ :=getString(args, "type")
	if dataType == "" {
		return err("type argument is required")
}

	var value string
	switch dataType {
	case "stepCount":
		value = "12345"
	case "heartRate":
		value = "72 bpm"
	case "activeEnergyBurned":
		value = "500 kcal"
	case "distanceWalkingRunning":
		value = "3.2 km"
	default:
		value = "unknown type"
	}
	return ok("Data for " + dataType + ": " + value)
}