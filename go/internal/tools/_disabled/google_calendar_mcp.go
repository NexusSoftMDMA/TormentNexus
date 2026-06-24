package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleListEvents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("GOOGLE_CALENDAR_TOKEN")
	calendarID, _ :=getString(args, "calendarId")
	if calendarID == "" {
		calendarID = "primary"
	}
	url := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/%s/events", calendarID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok("events: " + string(data))
}

func HandleCreateEvent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("GOOGLE_CALENDAR_TOKEN")
	calendarID, _ :=getString(args, "calendarId")
	if calendarID == "" {
		calendarID = "primary"
	}
	event := map[string]interface{}{
		"summary": getString(args, "summary"),
		"start":   map[string]string{"dateTime": getString(args, "startDateTime")},
		"end":     map[string]string{"dateTime": getString(args, "endDateTime")},
	}
	body, _ := json.Marshal(event)
	url := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/%s/events", calendarID)
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
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
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok("event created")
}