package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	phone, _ :=getString(args, "phone")
	message, _ :=getString(args, "message")
	if phone == "" || message == "" {
		return err("phone and message are required")
}

	body, _ := json.Marshal(map[string]string{"phone": phone, "message": message})
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "https://whatsapp-api.example.com/send", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
}

	return ok("message sent successfully")
}

func HandleListMessages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chatID, _ :=getString(args, "chat_id")
	url := "https://whatsapp-api.example.com/messages"
	if chatID != "" {
		url = url + "?chat_id=" + chatID
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(respBody))
}