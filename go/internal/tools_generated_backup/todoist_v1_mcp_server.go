package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("TODOIST_API_TOKEN")
	if token == "" {
		return err("TODOIST_API_TOKEN environment variable not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.todoist.com/rest/v1/projects", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	var projects []interface{}
	if e := json.Unmarshal(body, &projects); e != nil {
		return err(fmt.Sprintf("parse response: %v", e))
}

	return ok(fmt.Sprintf("Found %d projects", len(projects)))
}

func HandleCreateTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("TODOIST_API_TOKEN")
	if token == "" {
		return err("TODOIST_API_TOKEN environment variable not set")
}

	content, _ :=getString(args, "content")
	if content == "" {
		return err("'content' argument is required")
}

	payload := map[string]string{"content": content}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("marshal payload: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.todoist.com/rest/v1/tasks", bytes.NewReader(bodyBytes))
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return success("Task created successfully")
}