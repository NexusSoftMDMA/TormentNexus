package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListMeetings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userId, _ :=getString(args, "user_id")
	accessToken, _ :=getString(args, "access_token")
	if userId == "" || accessToken == "" {
		return err("user_id and access_token required")
}

	url := fmt.Sprintf("https://api.zoom.us/v2/users/%s/meetings", userId)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch meetings: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	return ok(fmt.Sprintf("Meetings: %s", string(body)))
}

func HandleGetUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userId, _ :=getString(args, "user_id")
	accessToken, _ :=getString(args, "access_token")
	if userId == "" || accessToken == "" {
		return err("user_id and access_token required")
}

	url := fmt.Sprintf("https://api.zoom.us/v2/users/%s", userId)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch user: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(fmt.Sprintf("User: %s", string(body)))
}