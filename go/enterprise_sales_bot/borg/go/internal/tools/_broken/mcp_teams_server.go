package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleListTeams_mcp_teams_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "accessToken")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://graph.microsoft.com/v1.0/me/joinedTeams", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}

func HandleSendMessage_mcp_teams_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	teamID, _ :=getString(args, "teamId")
	channelID, _ :=getString(args, "channelId")
	message, _ :=getString(args, "message")
	if teamID == "" || channelID == "" || message == "" {
		return err("teamId, channelId, and message are required")
}

	url := "https://graph.microsoft.com/v1.0/teams/" + teamID + "/channels/" + channelID + "/messages"
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+getString(args, "accessToken"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send message")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}