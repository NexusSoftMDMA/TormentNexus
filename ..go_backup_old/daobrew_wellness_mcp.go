package tools

import (
	"context"
	"fmt"
)

func HandleDetectStress(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	heartRate, _ :=getInt(args, "heartRate")
	sleepHours, _ :=getInt(args, "sleepHours")
	if heartRate > 100 && sleepHours < 6 {
		return success("High stress detected: elevated heart rate and insufficient sleep")
}

	if heartRate > 80 || sleepHours < 7 {
		return success("Moderate stress: consider relaxation techniques")
}

	return success("Low stress: biometrics are within healthy range")
}

func HandleGetRecovery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	stressLevel, _ :=getString(args, "stressLevel")
	duration, _ :=getInt(args, "durationMinutes")
	if stressLevel == "" {
		return err("Missing required parameter: stressLevel")
}

	plan := fmt.Sprintf("Recovery plan for %s stress over %d minutes: deep breathing, hydration, short walk", stressLevel, duration)
	return success(plan)
}