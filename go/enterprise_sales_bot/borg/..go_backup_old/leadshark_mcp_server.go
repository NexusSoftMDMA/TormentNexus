package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func HandleSendInvitation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	profileName, _ :=getString(args, "profileName")
	message, _ :=getString(args, "message")
	if profileName == "" {
		return err("profileName is required")
	}
	body := map[string]string{"profileName": profileName, "message": message}
	data, _ := json.Marshal(body)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.leadshark.ai/invite", bytes.NewBuffer(data))
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("API error: %s", resp.Status))
	}
	return ok("Invitation sent to " + profileName)
}

func HandleGetConnections(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.leadshark.ai/connections?limit=%d", limit), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("API error: %s", resp.Status))
	}
	return success("Fetched connections")
}