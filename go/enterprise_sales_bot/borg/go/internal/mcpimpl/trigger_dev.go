package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleTrigger(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	token, _ :=getString(args, "token")
	event, _ :=getString(args, "event")
	payload, _ :=getString(args, "payload")

	if url == "" || token == "" || event == "" {
		return err("missing required parameters: url, token, event")
}

	body := map[string]string{
		"event":   event,
		"payload": payload,
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal body: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return err("trigger returned status " + resp.Status)
}

	return success("trigger sent successfully")
}