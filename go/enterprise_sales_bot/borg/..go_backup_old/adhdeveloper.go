package tools

import (
	"context"
	"fmt"
)

func HandleBreakdownTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	task, _ :=getString(args, "task")
	if task == "" {
		return err("task is required")
}

	subtasks := fmt.Sprintf("- %s: step 1\n- %s: step 2\n- %s: step 3", task, task, task)
	return success(subtasks)
}

func HandlePomodoro(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	minutes, _ :=getInt(args, "minutes")
	if minutes <= 0 {
		return err("minutes must be positive")
}

	msg := fmt.Sprintf("Focus session started for %d minutes. Stay on task!", minutes)
	return ok(msg)
}