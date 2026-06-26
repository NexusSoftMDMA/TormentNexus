package tools

import (
	"context"
	"net/http"
)

func HandleCreateTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	taskID, _ :=getString(args, "task_id")
	if taskID == "" {
		return err("task_id is required")
}

	_, e := http.DefaultClient.Get("https://example.com/create?task=" + taskID)
	if e != nil {
		return err("failed to create task: " + e.Error())
}

	return ok("task created")
}

func HandleGetTaskStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	taskID, _ :=getString(args, "task_id")
	if taskID == "" {
		return err("task_id is required")
}

	_, e := http.DefaultClient.Get("https://example.com/status?task=" + taskID)
	if e != nil {
		return err("failed to get status: " + e.Error())
}

	return ok("task is pending")
}