package mcpimpl

import (
	"context"
	"fmt"
)

var tasks = make(map[string]string)

func HandleSubmitTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	desc, _ :=getString(args, "description")
	if desc == "" {
		return err("description is required")
}

	id := fmt.Sprintf("task-%d", len(tasks)+1)
	tasks[id] = desc
	return success("Task submitted with ID: " + id)
}

func HandleGetTaskStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	desc, found := tasks[id]
	if !found {
		return err("task not found")
}

	return ok(desc)
}