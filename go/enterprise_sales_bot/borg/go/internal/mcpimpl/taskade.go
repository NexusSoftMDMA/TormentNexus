package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateTask_taskade(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	projectID, _ :=getString(args, "project_id")
	if title == "" || projectID == "" {
		return err("title and project_id are required")
}

	body, _ := json.Marshal(map[string]string{"title": title, "project_id": projectID})
	resp, e := http.DefaultClient.Post("https://api.taskade.com/v1/tasks", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to create task: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("taskade API returned status " + resp.Status)
}

	return ok("Task created successfully")
}

func HandleListTasks_taskade(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	if projectID == "" {
		return err("project_id is required")
}

	url := fmt.Sprintf("https://api.taskade.com/v1/projects/%s/tasks", projectID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to list tasks: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("taskade API returned status " + resp.Status)
}

	var result interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return ok("Tasks listed successfully")
}