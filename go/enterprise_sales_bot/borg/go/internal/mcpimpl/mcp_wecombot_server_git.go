package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSendText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	webhook, _ :=getString(args, "webhook_url")
	content, _ :=getString(args, "content")
	if webhook == "" || content == "" {
		return err("webhook_url and content are required")
}

	body, _ := json.Marshal(map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": content},
	})
	resp, e := http.DefaultClient.Post(webhook, "application/json", bytes.NewBuffer(body))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return ok("text sent successfully")
}

func HandleSendMarkdown(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	webhook, _ :=getString(args, "webhook_url")
	content, _ :=getString(args, "content")
	if webhook == "" || content == "" {
		return err("webhook_url and content are required")
}

	body, _ := json.Marshal(map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{"content": content},
	})
	resp, e := http.DefaultClient.Post(webhook, "application/json", bytes.NewBuffer(body))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return ok("markdown sent successfully")
}