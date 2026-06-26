package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleHyperchatPostMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	channel, _ :=getString(args, "channel")
	message, _ :=getString(args, "message")
	if channel == "" || message == "" {
		return err("channel and message are required")
}

	url := "https://api.hyperchat.example.com/messages"
	payload := map[string]string{"channel": channel, "message": message}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request")
}

	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewBuffer(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBytes, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s", string(respBytes)))
}

	return ok("message sent successfully")
}