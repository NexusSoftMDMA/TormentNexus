package tools

import (
	"bytes"
	"context"
	"net/http"
)

func HandleListConnectors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://graph.microsoft.com/v1.0/connections", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	return ok("Connectors listed")
}

func HandleSendNotification(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	webhookUrl, _ :=getString(args, "webhookUrl")
	message, _ :=getString(args, "message")
	if webhookUrl == "" || message == "" {
		return err("webhookUrl and message are required")
}

	body := []byte(`{"text":"` + message + `"}`)
	req, e := http.NewRequestWithContext(ctx, "POST", webhookUrl, bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	return ok("Notification sent (status: " + resp.Status + ")")
}