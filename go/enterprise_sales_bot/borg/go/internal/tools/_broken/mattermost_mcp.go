package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleListChannels_mattermost_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL := os.Getenv("MATTERMOST_URL")
	token := os.Getenv("MATTERMOST_TOKEN")
	if serverURL == "" || token == "" {
		return err("missing Mattermost URL or token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", serverURL+"/api/v4/channels", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	return ok(fmt.Sprintf("Channels: %v", result))
}

func HandlePostMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL := os.Getenv("MATTERMOST_URL")
	token := os.Getenv("MATTERMOST_TOKEN")
	channelID, _ :=getString(args, "channel_id")
	message, _ :=getString(args, "message")
	if serverURL == "" || token == "" || channelID == "" || message == "" {
		return err("missing required parameters")
}

	body := map[string]string{"channel_id": channelID, "message": message}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", serverURL+"/api/v4/posts", nil)
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
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	return ok(fmt.Sprintf("Posted: %v", result))
}