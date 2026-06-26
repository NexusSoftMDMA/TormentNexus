package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func sendSMS(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	to, _ :=getString(args, "to")
	message, _ :=getString(args, "message")
	if apiKey == "" || to == "" || message == "" {
		return err("missing required fields: api_key, to, message")
}

	body, _ := json.Marshal(map[string]string{
		"api_key": apiKey,
		"to":      to,
		"message": message,
	})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.didlogic.com/v1/sms/send/", bytes.NewReader(body))
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
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	if id, found := result["message_id"]; found {
		return ok(fmt.Sprintf("SMS sent, ID: %v", id))
}

	return err("unexpected response")
}

func getBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("missing api_key")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.didlogic.com/v1/account/balance?api_key=%s", apiKey), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	if balance, found := result["balance"]; found {
		return ok(fmt.Sprintf("Balance: %v", balance))
}

	return err("unexpected response")
}