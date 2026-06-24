package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListCalendars(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.keeper.sh/calendars")
	if e != nil {
		return err("failed to fetch calendars: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok("calendars retrieved successfully")
}

func HandleSyncCalendar(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	calendarID, _ :=getString(args, "calendarId")
	if calendarID == "" {
		return err("calendarId is required")
	}
	url := "https://api.keeper.sh/calendars/" + calendarID + "/sync"
	resp, e := http.DefaultClient.Post(url, "application/json", nil)
	if e != nil {
		return err("sync request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return success("calendar synced successfully")
}