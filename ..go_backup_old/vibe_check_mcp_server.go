package tools

import "context"

func HandleVibeCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	task, _ :=getString(args, "task")
	vibe := "Vibe check passed. Focus on simplicity and avoid over-engineering."
	message := "Task: " + task + ". " + vibe
	return ok(message)
}

func HandleAnalyzeRisk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	riskLevel, _ :=getString(args, "risk_level")
	feedback := "Risk level '" + riskLevel + "' detected. Consider breaking down the task."
	return success(feedback)
}