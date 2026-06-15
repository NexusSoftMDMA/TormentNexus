package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleGetTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "https://api.example.com/tasks"
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	var tasks interface{}
	if e := json.NewDecoder(resp.Body).Decode(&tasks); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	return ok(fmt.Sprintf("tasks: %v", tasks))
}

func HandleCreateTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	if title == "" {
		return err("title is required")
}

	url, _ :=getString(args, "url")
	if url == "" {
		url = "https://api.example.com/tasks"
	}
	body := strings.NewReader(fmt.Sprintf(`{"title":"%s"}`, title))
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return success("task created")
}