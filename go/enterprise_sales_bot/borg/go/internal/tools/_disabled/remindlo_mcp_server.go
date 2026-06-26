package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleSendReminder(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	phone, _ :=getString(args, "phone")
	message, _ :=getString(args, "message")
	if phone == "" || message == "" {
		return err("phone and message are required")
	}
	body, e := json.Marshal(map[string]string{"phone": phone, "message": message})
	if e != nil {
		return err("failed to marshal request: " + e.Error())
	}
	resp, e := http.DefaultClient.Post("https://api.remindlo.com/reminders", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return err("API returned status " + resp.Status)
	}
	return ok("Reminder sent successfully")
}