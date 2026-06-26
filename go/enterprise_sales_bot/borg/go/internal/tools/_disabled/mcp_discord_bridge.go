package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleSendDiscordMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	webhookURL, _ :=getString(args, "webhook_url")
	content, _ :=getString(args, "content")
	if webhookURL == "" || content == "" {
		return err("webhook_url and content are required")
	}
	payload := map[string]string{"content": content}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
	}
	resp, e := http.DefaultClient.Post(webhookURL, "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to send message: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return err("Discord returned status " + resp.Status)
	}
	return ok("Message sent successfully")
}