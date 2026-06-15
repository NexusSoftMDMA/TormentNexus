package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListEvents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	calID, _ :=getString(args, "calendarId")
	url := fmt.Sprintf("https://api.cal.com/v1/events?calendarId=%s", calID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch events")
}

	defer resp.Body.Close()
	var events []map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&events); e != nil {
		return err("invalid event data")
}

	return ok(fmt.Sprintf("found %d events", len(events)))
}

func HandleCreateEvent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	body := map[string]interface{}{
		"title":     getString(args, "title"),
		"startTime": getString(args, "startTime"),
		"endTime":   getString(args, "endTime"),
	}
	jsonData, e := json.Marshal(body)
	if e != nil {
		return err("failed to serialize request")
}

	resp, e := http.DefaultClient.Post("https://api.cal.com/v1/events", "application/json", bytes.NewReader(jsonData))
	if e != nil {
		return err("failed to create event")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err("event creation failed")
}

	return success("event created")
}