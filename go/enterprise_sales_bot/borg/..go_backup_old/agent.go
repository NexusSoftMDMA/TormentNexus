package tools

import "context"

func HandleAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	task, _ :=getString(args, "task")
	msg := "Agent executed task: " + task
	return success(msg)
}