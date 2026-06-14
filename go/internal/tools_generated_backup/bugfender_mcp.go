package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// handleListApps returns all apps from Bugfender.
func handleListApps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.bugfender.com/v1/apps", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %s %s", resp.Status, string(body)))
}

	return ok(string(body))
}

// handleListLogs returns logs for a specific Bugfender app.
func handleListLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token is required")
}

	appID, _ :=getString(args, "appId")
	if appID == "" {
		return err("appId is required")
}

	url := fmt.Sprintf("https://api.bugfender.com/v1/apps/%s/logs", appID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %s %s", resp.Status, string(body)))
}

	return ok(string(body))
}