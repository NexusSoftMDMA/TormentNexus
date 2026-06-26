package tools

import (
	"context"
	"fmt"
)

// HandleLogWaterIntake logs water intake.
func HandleLogWaterIntake(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	amount, _ :=getInt(args, "amount")
	unit, _ :=getString(args, "unit")
	if unit == "" {
		unit = "ml"
	}
	return ok(fmt.Sprintf("Logged %d %s water", amount, unit))
}

// HandleGetDailyHydration returns daily hydration summary.
func HandleGetDailyHydration(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	total := 0 // placeholder
	goal := 2000
	msg := fmt.Sprintf("Hydration total: %d ml / goal: %d ml", total, goal)
	return success(msg)
}