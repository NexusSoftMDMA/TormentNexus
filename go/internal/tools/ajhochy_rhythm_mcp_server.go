//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleListTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := "http://localhost:8080"
	resp, e := http.DefaultClient.Get(baseURL + "/tasks")
	if e != nil {
		return err("failed to fetch tasks: " + e.Error())
}

	defer resp.Body.Close()
	var tasks []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&tasks); e != nil {
		return err("failed to decode tasks: " + e.Error())
}

	data, e := json.Marshal(tasks)
	if e != nil {
		return err("failed to marshal tasks: " + e.Error())
}

	return ok(string(data))
}

func HandleCreateTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	project, _ :=getString(args, "project")
	if title == "" {
		return err("title is required")
}

	body := map[string]string{"title": title, "project": project}
	payload, _ := json.Marshal(body)
	baseURL := "http://localhost:8080"
	resp, e := http.DefaultClient.Post(baseURL+"/tasks", "application/json", bytes.NewReader(payload))
	if e != nil {
		return err("failed to create task: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return err("failed to create task: status " + resp.Status)
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}
