package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type calendarEvent struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
	Start   struct {
		DateTime string `json:"dateTime"`
		Date     string `json:"date"`
	} `json:"start"`
	End struct {
		DateTime string `json:"dateTime"`
		Date     string `json:"date"`
	} `json:"end"`
}

type calendarListResponse struct {
	Items []calendarEvent `json:"items"`
}

func HandleListEvents_mcp_google_calendar(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	calendarID, _ :=getString(args, "calendarId")
	accessToken, _ :=getString(args, "accessToken")
	if calendarID == "" || accessToken == "" {
		return err("calendarId and accessToken are required")
}

	url := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/%s/events", calendarID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result calendarListResponse
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	out, _ := json.Marshal(result.Items)
	return success(string(out))
}