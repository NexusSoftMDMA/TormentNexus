package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func HandleMattermostSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	token, _ :=getString(args, "token")
	channelID, _ :=getString(args, "channel_id")
	message, _ :=getString(args, "message")
	if serverURL == "" || token == "" || channelID == "" || message == "" {
		return err("missing required arguments: server_url, token, channel_id, message")
}

	body, e := json.Marshal(map[string]string{"channel_id": channelID, "message": message})
	if e != nil {
		return err("failed to marshal request body: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(serverURL, "/")+"/api/v4/posts", strings.NewReader(string(body)))
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
	if resp.StatusCode != http.StatusCreated {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return err(fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, string(respBody)))
}

	return ok("message sent successfully")
}