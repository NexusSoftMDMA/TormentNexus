package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleCreateTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	priority, _ :=getInt(args, "priority")
	body := map[string]interface{}{
		"title":    title,
		"content":  content,
		"priority": priority,
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("Failed to marshal request body")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.ticktick.com/open/v1/task", strings.NewReader(string(jsonBody)))
	if e != nil {
		return err("Failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Failed to send request: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return err("Unexpected status: " + resp.Status)
}

	return ok("Task created successfully")
}

func HandleListTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.ticktick.com/open/v1/task", nil)
	if e != nil {
		return err("Failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Failed to fetch tasks")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("Unexpected status: " + resp.Status)
}

	return ok("Tasks retrieved successfully")
}