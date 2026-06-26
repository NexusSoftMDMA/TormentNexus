package mcpimpl

import (
	"context"
)

func HandleGetAtmosphere(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	if city == "" {
		return err("city is required")
}

	result := map[string]interface{}{
		"city":        city,
		"temperature": 22.5,
		"humidity":    65,
		"description": "Partly cloudy",
	}
	return ok(result)
}