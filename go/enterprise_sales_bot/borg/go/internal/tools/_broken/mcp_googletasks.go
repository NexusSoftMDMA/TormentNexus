package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListTasks_mcp_googletasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://tasks.googleapis.com/tasks/v1/lists/@default/tasks", nil)
	if e != nil {
		return err("create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer " + token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body: " + e.Error())
}

	return ok(string(body))
}

func HandleAddTask_mcp_googletasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	title, _ :=getString(args, "title")
	if token == "" || title == "" {
		return err("token and title required")
}

	payload := map[string]string{"title": title}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("marshal: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://tasks.googleapis.com/tasks/v1/lists/@default/tasks", io.NopCloser(body))
	if e != nil {
		return err("create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer " + token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response: " + e.Error())
}

	return ok(string(respBody))
}