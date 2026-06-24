package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateJiraIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	summary, _ :=getString(args, "summary")
	description, _ :=getString(args, "description")
	if project == "" || summary == "" {
		return err("project and summary are required")
}

	body := map[string]interface{}{
		"fields": map[string]interface{}{
			"project":     map[string]interface{}{"key": project},
			"summary":     summary,
			"description": description,
		},
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://your-jira-instance/rest/api/2/issue", bytes.NewBuffer(jsonBody))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
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

	return ok("Issue created successfully")
}

func HandleSendSlackMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	channel, _ :=getString(args, "channel")
	text, _ :=getString(args, "text")
	if channel == "" || text == "" {
		return err("channel and text are required")
}

	body := map[string]interface{}{
		"channel": channel,
		"text":    text,
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(jsonBody))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return ok("Message sent successfully")
}