package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateMeeting(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	startTime, _ :=getString(args, "start_time")
	duration, _ :=getInt(args, "duration")
	token, _ :=getString(args, "access_token")
	if topic == "" || startTime == "" || duration == 0 || token == "" {
		return err("missing required arguments: topic, start_time, duration, access_token")
}

	body := map[string]interface{}{
		"topic":      topic,
		"start_time": startTime,
		"duration":   duration,
		"type":       2,
	}
	b, e := json.Marshal(body)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.zoom.us/v2/users/me/meetings", bytes.NewReader(b))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err(fmt.Sprintf("zoom api error: status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	return success(fmt.Sprintf("meeting created: %v", result["id"]))
}

func HandleListMeetings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	if token == "" {
		return err("missing access_token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.zoom.us/v2/users/me/meetings?type=scheduled", nil)
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("zoom api error: status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	meetings, found := result["meetings"].([]interface{})
	if !found {
		return err("unexpected response format")
}

	out, e := json.MarshalIndent(meetings, "", "  ")
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	return success(string(out))
}