package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	status, _ :=getString(args, "status")
	url := "https://api.idealift.com/tasks"
	if status != "" {
		url += "?status=" + status
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch tasks")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON")
}

	return ok(fmt.Sprintf("Tasks: %v", result))
}

func HandleCreateTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	if title == "" {
		return err("title is required")
}

	desc, _ :=getString(args, "description")
	payload := map[string]string{"title": title, "description": desc}
	data, e := json.Marshal(payload)
	if e != nil {
		return err("failed to encode")
}

	resp, e := http.DefaultClient.Post("https://api.idealift.com/tasks", "application/json", bytes.NewReader(data))
	if e != nil {
		return err("failed to create task")
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err("failed to create task")
}

	return success("task created")
}