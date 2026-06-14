package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleCallAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "agent_url")
	if url == "" {
		return err("agent_url is required")
}

	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	payload := map[string]string{"message": message}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(result)
}