package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	desc, _ :=getString(args, "description")
	payload := map[string]string{"title": title, "description": desc}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.taskmaster.ai/tasks", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return ok(fmt.Sprintf("task created: %v", result["id"]))
}

func HandleListTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.taskmaster.ai/tasks", nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var tasks []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&tasks); e != nil {
		return err("failed to decode response")
}

	data, e := json.MarshalIndent(tasks, "", "  ")
	if e != nil {
		return err("failed to marshal tasks")
}

	return ok(string(data))
}