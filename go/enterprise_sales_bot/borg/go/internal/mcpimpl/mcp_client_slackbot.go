package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleSendSlackMessage_mcp_client_slackbot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	webhook, _ :=getString(args, "webhook_url")
	message, _ :=getString(args, "message")
	if webhook == "" || message == "" {
		return err("webhook_url and message are required")
}

	payload, e := json.Marshal(map[string]string{"text": message})
	if e != nil {
		return err("failed to marshal message: " + e.Error())
}

	resp, e := http.DefaultClient.Post(webhook, "application/json", bytes.NewBuffer(payload))
	if e != nil {
		return err("failed to send message: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("slack returned status: " + resp.Status)
}

	return ok("message sent")
}