package tools

import (
	"context"
	"net/http"
	"strconv"
)

func HandleSearchTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.backlog.com/tasks?q="+query, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	return success("search completed for: " + query)
}

func HandleGetTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "task_id")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.backlog.com/tasks/"+strconv.Itoa(id), nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	return success("task retrieved: " + strconv.Itoa(id))
}