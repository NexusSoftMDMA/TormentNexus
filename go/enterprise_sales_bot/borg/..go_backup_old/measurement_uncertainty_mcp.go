package tools

import "context"

func HandleMeasurementUncertainty(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	value, _ :=getString(args, "value")
	uncertainty, _ :=getString(args, "uncertainty")
	if value == "" {
		return err("value is required")
}

	if uncertainty == "" {
		return err("uncertainty is required")
}

	return ok("value: " + value + " ± " + uncertainty)
}