package tools

import (
	"context"
	"io"
	"net/http"
	"strconv"
)

func HandleListEvents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	calendarId, _ :=getString(args, "calendarId")
	if calendarId == "" {
		calendarId = "primary"
	}
	maxResults, _ :=getInt(args, "maxResults")
	if maxResults <= 0 {
		maxResults = 10
	}
	url := "https://www.googleapis.com/calendar/v3/calendars/" + calendarId + "/events?maxResults=" + strconv.Itoa(maxResults)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch events: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleListCalendars(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://www.googleapis.com/calendar/v3/users/me/calendarList")
	if e != nil {
		return err("failed to list calendars: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}