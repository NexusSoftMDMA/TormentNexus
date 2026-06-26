package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleGetEvent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	eventId, _ :=getString(args, "eventId")
	if eventId == "" {
		return err("eventId is required")
}

	calendarId, _ :=getString(args, "calendarId")
	if calendarId == "" {
		calendarId = "primary"
	}
	url := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/%s/events/%s", calendarId, eventId)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + string(body))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Event: %v", result))
}