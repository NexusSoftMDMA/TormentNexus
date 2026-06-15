package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleSendNotification(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("Missing api_key")
}

	message, _ :=getString(args, "message")
	if message == "" {
		return err("Missing message")
}

	userId, _ :=getString(args, "user_id")

	body := map[string]interface{}{
		"message": message,
		"user_id": userId,
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("Failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.gotohuman.com/v1/notifications", bytes.NewBuffer(jsonBody))
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return err("API returned status " + resp.Status)
}

	return ok("Notification sent")
}