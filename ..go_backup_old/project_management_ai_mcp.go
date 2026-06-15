package tools

import "context"

func HandleDecomposeTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	task, _ :=getString(args, "task")
	return ok("Decomposed task: " + task)
}

func HandlePlanSprint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sprint, _ :=getString(args, "sprint")
	return ok("Planned sprint: " + sprint)
}