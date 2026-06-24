package tools

import (
	"context"
	"encoding/json"
)

func HandleGetSleep(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	date, _ :=getString(args, "date")
	sleepData := map[string]interface{}{
		"date":     date,
		"duration": 7.5,
		"quality":  "good",
	}
	raw, e := json.Marshal(sleepData)
	if e != nil {
		return err("failed to marshal sleep data")
}

	return ok(string(raw))
}

func HandleGetSteps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	date, _ :=getString(args, "date")
	stepsData := map[string]interface{}{
		"date":  date,
		"steps": 8432,
	}
	raw, e := json.Marshal(stepsData)
	if e != nil {
		return err("failed to marshal steps data")
}

	return ok(string(raw))
}