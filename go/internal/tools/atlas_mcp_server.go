//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreateTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	body := map[string]interface{}{
		"name":        getString(args, "name"),
		"description": getString(args, "description"),
		"projectId":   getString(args, "projectId"),
	}
	data, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:8080/tasks", bytes.NewReader(data))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("unexpected status: %s - %s", resp.Status, string(b)))
}

	return ok("task created")
}

func HandleGetTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "taskId")
	if id == "" {
		return err("taskId is required")
}

	url := fmt.Sprintf("http://localhost:8080/tasks/%s", id)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("unexpected status: %s - %s", resp.Status, string(b)))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return success(result)
}
