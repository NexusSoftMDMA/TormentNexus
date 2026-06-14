package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleSendNotification(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	user, _ :=getString(args, "user")
	message, _ :=getString(args, "message")
	if token == "" || user == "" || message == "" {
		return err("token, user, and message are required")
}

	payload := map[string]interface{}{
		"token":   token,
		"user":    user,
		"message": message,
	}
	if title := getString(args, "title"); title != "" {
		payload["title"] = title
	}
	if priority := getInt(args, "priority"); priority != 0 {
		payload["priority"] = priority
	}
	if sound := getString(args, "sound"); sound != "" {
		payload["sound"] = sound
	}
	if device := getString(args, "device"); device != "" {
		payload["device"] = device
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://api.pushover.net/1/messages.json", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("pushover error: " + resp.Status)
}

	status, found := result["status"].(float64)
	if !found || status != 1 {
		msg, _ := result["errors"].([]interface{})
		return err("pushover rejection: " + formatErrors(msg))
}

	receipt, _ := result["receipt"].(string)
	return ok("notification sent, receipt: " + receipt)
}

func formatErrors(errors []interface{}) string {
	if len(errors) == 0 {
		return "unknown error"
	}
	return errors[0].(string)
}