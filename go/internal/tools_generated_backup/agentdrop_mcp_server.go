package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleAgentDropSend(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	message, _ :=getString(args, "message")
	body, e := json.Marshal(map[string]string{"message": message})
	if e != nil {
		return err("failed to marshal message")
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	return ok("message sent")
}

func HandleAgentDropReceive(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	var result interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to parse response")
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}