package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetDepartures(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	station, _ :=getString(args, "station_code")
	if station == "" {
		return err("station_code is required")
}

	count, _ :=getInt(args, "count", 10)

	url := fmt.Sprintf("https://api.nationalrail.co.uk/departures/%s?count=%d", station, count)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("Departures: %v", data))
}