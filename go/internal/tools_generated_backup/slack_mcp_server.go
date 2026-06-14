package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		return err("SLACK_TOKEN not set")
}

	channel, _ :=getString(args, "channel")
	if channel == "" {
		return err("missing channel")
}

	text, _ :=getString(args, "text")
	if text == "" {
		return err("missing text")
}

	payload := map[string]string{"channel": channel, "text": text}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/chat.postMessage", bytes.NewReader(body))
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
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	if found, _ := result["ok"].(bool); !found {
		return err("slack api error: " + fmt.Sprint(result))
}

	return ok("message sent")
}