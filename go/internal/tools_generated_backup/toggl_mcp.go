package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetMe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("TOGGL_API_TOKEN")
	if token == "" {
		return err("missing TOGGL_API_TOKEN")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.track.toggl.com/api/v9/me", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.SetBasicAuth(token, "api_token")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse response: %v", e))
}

	return ok(fmt.Sprintf("User: %v", result["fullname"]))
}

func HandleStartTimeEntry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("TOGGL_API_TOKEN")
	if token == "" {
		return err("missing TOGGL_API_TOKEN")
}

	description, _ :=getString(args, "description")
	if description == "" {
		return err("description is required")
}

	payload := map[string]interface{}{
		"time_entry": map[string]interface{}{
			"description": description,
			"created_with": "mcp",
		},
	}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("marshal payload: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.track.toggl.com/api/v9/time_entries", nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.SetBasicAuth(token, "api_token")
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
}

	return ok(fmt.Sprintf("Started time entry: %s", description))
}