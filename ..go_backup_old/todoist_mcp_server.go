package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "apiToken")
	if token == "" {
		return err("apiToken is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.todoist.com/rest/v2/tasks", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	var tasks []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&tasks); e != nil {
		return err(fmt.Sprintf("failed to decode: %v", e))
}

	out, _ := json.Marshal(tasks)
	return ok(string(out))
}

func HandleCreateTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "apiToken")
	content, _ :=getString(args, "content")
	if token == "" || content == "" {
		return err("apiToken and content are required")
}

	body, _ := json.Marshal(map[string]string{"content": content})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.todoist.com/rest/v2/tasks", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	var task map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&task); e != nil {
		return err(fmt.Sprintf("failed to decode: %v", e))
}

	out, _ := json.Marshal(task)
	return ok(string(out))
}