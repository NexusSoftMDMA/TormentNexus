package tools

import (
	"context"
	"os/exec"
)

func HandleListTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("task", "list")
	output, e := cmd.Output()
	if e != nil {
		return err("failed to run task: " + e.Error())
}

	return success(string(output))
}

func HandleAddTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	description, _ :=getString(args, "description")
	if description == "" {
		return err("description is required")
}

	cmd := exec.Command("task", "add", description)
	output, e := cmd.Output()
	if e != nil {
		return err("failed to add task: " + e.Error())
}

	return success(string(output))
}