package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetFlight(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	flightID, _ :=getString(args, "flight_id")
	if flightID == "" {
		return err("flight_id is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://api.flightradar24.com/common/v1/flight?flight_id=%s", flightID))
	if e != nil {
		return err("failed to fetch flight: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Flight data: %v", data))
}

func HandleSearchFlights(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	airline, _ :=getString(args, "airline")
	if airline == "" {
		return err("airline is required")
}

	url := fmt.Sprintf("https://api.flightradar24.com/common/v1/search?query=%s&limit=10", airline)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("search failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	var data []map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("JSON parse error: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d flights", len(data)))
}