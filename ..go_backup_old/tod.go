package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleGetTodos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	todos := []map[string]interface{}{
		{"id": "1", "task": "Learn MCP", "done": false},
		{"id": "2", "task": "Build Tod server", "done": true},
	}
	data, e := json.Marshal(todos)
	if e != nil {
		return err("failed to marshal todos")
}

	return ok(string(data))
}

func HandleAddTodo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	task, _ :=getString(args, "task")
	if task == "" {
		return err("task is required")
}

	return success(fmt.Sprintf("Added todo: %s", task))
}