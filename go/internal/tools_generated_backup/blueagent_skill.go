package tools

import (
	"context"
)

func HandleBlueIdea(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	return success("blue_idea executed for: " + prompt)
}

func HandleBlueBuild(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	task, _ :=getString(args, "task")
	if task == "" {
		return err("task is required")
}

	return ok("blue_build started: " + task)
}