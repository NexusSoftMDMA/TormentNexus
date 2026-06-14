package tools

import (
	"context"
	"fmt"
)

func HandleListDidaTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tasks := []string{"Grocery shopping", "Finish report", "Call Mom"}
	return ok(fmt.Sprintf("Tasks: %v", tasks))
}

func HandleCreateDidaTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	if title == "" {
		return err("title is required")
}

	return success(fmt.Sprintf("Task '%s' created", title))
}