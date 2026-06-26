package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleTimetable(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	station, _ :=getString(args, "station")
	if station == "" {
		return err("station is required")
}

	url := fmt.Sprintf("https://api.timetable.com/v1/timetable?station=%s", station)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch timetable: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return success(fmt.Sprintf("Timetable data: %v", result))
}