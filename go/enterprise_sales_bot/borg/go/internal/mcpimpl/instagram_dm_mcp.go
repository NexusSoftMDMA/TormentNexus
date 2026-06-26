package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleSendDm(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "userId")
	message, _ :=getString(args, "message")
	body, _ := json.Marshal(map[string]string{"recipient": userID, "message": message})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://graph.instagram.com/v12.0/me/messages", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	return ok("DM sent to " + userID)
}

func HandleGetDms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "userId")
	limit, _ :=getInt(args, "limit")
	url := "https://graph.instagram.com/v12.0/me/conversations?user_id=" + userID + "&limit=" + string(rune(limit))
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	return ok("DMs fetched")
}