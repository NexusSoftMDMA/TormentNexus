package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// HandleSendMessage sends a message to a Webex room.
func HandleSendMessage_webex_messaging_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	roomID, _ :=getString(args, "roomId")
	text, _ :=getString(args, "text")
	if token == "" || roomID == "" || text == "" {
		return err("missing required args: token, roomId, text")
}

	body, _ := json.Marshal(map[string]string{"roomId": roomID, "text": text})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.ciscospark.com/v1/messages", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err("unexpected status: " + resp.Status)
}

	var result struct{ ID string }
	json.NewDecoder(resp.Body).Decode(&result)
	return ok("message sent: " + result.ID)
}