package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	channel, _ :=getString(args, "channel")
	text, _ :=getString(args, "text")

	body := fmt.Sprintf(`{"channel":"%s","text":"%s"}`, channel, text)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/chat.postMessage", strings.NewReader(body))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request")
}

	defer resp.Body.Close()

	var result struct {
		Ok bool `json:"ok"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	if !result.Ok {
		return err("slack API returned error")
}

	return ok("message sent to " + channel)
}

func HandleListChannels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")

	req, e := http.NewRequestWithContext(ctx, "GET", "https://slack.com/api/conversations.list", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request")
}

	defer resp.Body.Close()

	var result struct {
		Ok       bool `json:"ok"`
		Channels []struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"channels"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	if !result.Ok {
		return err("slack API returned error")
}

	var names []string
	for _, c := range result.Channels {
		names = append(names, c.Name)

	return ok("channels: " + strings.Join(names, ", "))
}
}