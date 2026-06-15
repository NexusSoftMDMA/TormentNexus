package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleSendNotification(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	message, _ :=getString(args, "message")
	body := map[string]interface{}{
		"topic":   topic,
		"message": message,
	}
	if title := getString(args, "title"); title != "" {
		body["title"] = title
	}
	if priority := getString(args, "priority"); priority != "" {
		body["priority"] = priority
	}
	if tags := getString(args, "tags"); tags != "" {
		body["tags"] = tags
	}
	jsonData, e := json.Marshal(body)
	if e != nil {
		return err("failed to encode request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://ntfy.sh", bytes.NewReader(jsonData))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err("ntfy returned status: " + resp.Status)
}

	return ok("notification sent successfully")
}