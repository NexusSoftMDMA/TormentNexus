package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleEmlSend(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
	}
	body := map[string]string{"message": message}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal body")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.engramprotocol.com/v1/eml", bytes.NewReader(jsonBody))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
	}
	msg := "EML message sent"
	if v, found := result["message"]; found {
		msg = v.(string)

	return success(msg)
}
}