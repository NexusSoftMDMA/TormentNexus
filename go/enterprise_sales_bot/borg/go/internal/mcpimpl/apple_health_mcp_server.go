package mcpimpl

import (
	"context"
	"encoding/json"
)

func HandleGetSteps_apple_health_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	startDate, _ :=getString(args, "startDate")
	endDate, _ :=getString(args, "endDate")
	data := map[string]interface{}{
		"steps": 12345,
		"start": startDate,
		"end":   endDate,
	}
	jsonBytes, e := json.Marshal(data)
	if e != nil {
		return err("failed to marshal response")
}

	return ok(string(jsonBytes))
}

func HandleGetHeartRate_apple_health_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	startDate, _ :=getString(args, "startDate")
	endDate, _ :=getString(args, "endDate")
	data := map[string]interface{}{
		"averageHeartRate": 72,
		"start":            startDate,
		"end":              endDate,
	}
	jsonBytes, e := json.Marshal(data)
	if e != nil {
		return err("failed to marshal response")
}

	return ok(string(jsonBytes))
}