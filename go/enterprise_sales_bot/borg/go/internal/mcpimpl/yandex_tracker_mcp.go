package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetTask_yandex_tracker_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	taskID, _ :=getString(args, "task_id")
	if taskID == "" {
		return err("task_id is required")
}

	token, _ :=getString(args, "token")
	orgID, _ :=getString(args, "org_id")
	if token == "" || orgID == "" {
		return err("token and org_id are required")
}

	url := fmt.Sprintf("https://api.tracker.yandex.net/v2/issues/%s", taskID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "OAuth "+token)
	req.Header.Set("X-Org-ID", orgID)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response")
}

	summary, _ :=getString(data, "summary")
	description, _ :=getString(data, "description")
	return ok(fmt.Sprintf("Task %s: %s - %s", taskID, summary, description))
}