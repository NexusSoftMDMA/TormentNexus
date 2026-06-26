package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ntfyPayload struct {
	Topic    string   `json:"topic"` //nolint:unused
	Message  string   `json:"message"`
	Title    string   `json:"title,omitempty"`
	Priority int      `json:"priority,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

func HandleSendNtfy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		return err("topic is required")
}

	payload := ntfyPayload{
		Topic:    topic,
		Message:  getString(args, "message"),
		Title:    getString(args, "title"),
		Priority: getInt(args, "priority"),
	}
	tags, _ :=getString(args, "tags")
	if tags != "" {
		payload.Tags = []string{tags}
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://ntfy.sh/"+topic, "application/json", bytes.NewReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("ntfy returned status %d", resp.StatusCode))
}

	return ok("Notification sent successfully")
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong")
}