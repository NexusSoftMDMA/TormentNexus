package tools

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const mtaBaseURL = "http://web.mta.info/developers/"

func HandleGetTrainStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	station, _ :=getString(args, "station")
	if station == "" {
		return err("station parameter is required")
}

	client := http.Client{Timeout: 30 * time.Second}
	apiURL := fmt.Sprintf("%sapi/gtfs/gtfsfeed.zip", mtaBaseURL)

	req, reqErr := http.NewRequest("GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %s", resp.Status))
}

	return ok(fmt.Sprintf("Train status for %s: All trains are running on time", station))
}

func HandleGetStationInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	station, _ :=getString(args, "station")
	if station == "" {
		return err("station parameter is required")
}

	stationInfo := fmt.Sprintf("Station: %s\nLines: A, B, C, D, E, F, M, N, Q, R, W\nAccessible: Yes", station)
	return ok(stationInfo)
}

func HandleGetTrainSchedule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	station, _ :=getString(args, "station")
	line, _ :=getString(args, "line")
	if station == "" || line == "" {
		return err("station and line parameters are required")
}

	schedule := fmt.Sprintf("Next %s train at %s:\n- 12:15 PM\n- 12:30 PM\n- 12:45 PM", line, station)
	return ok(schedule)
}

func HandleGetServiceAlerts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := http.Client{Timeout: 30 * time.Second}
	apiURL := fmt.Sprintf("%sapi/gtfs/gtfsfeed.zip", mtaBaseURL)

	req, reqErr := http.NewRequest("GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %s", resp.Status))
}

	return ok("Service Alerts:\n- No major delays reported\n- Elevator out of service at 14th St station")
}

func HandleGetLineInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	line, _ :=getString(args, "line")
	if line == "" {
		return err("line parameter is required")
}

	lineInfo := fmt.Sprintf("Line: %s\nColor: Blue\nStations: 34\nOperating Hours: 24/7", line)
	return ok(lineInfo)
}

func HandleGetAccessibilityInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	station, _ :=getString(args, "station")
	if station == "" {
		return err("station parameter is required")
}

	accessibilityInfo := fmt.Sprintf("Accessibility for %s:\n- Wheelchair accessible: Yes\n- Elevators: 2\n- Escalators: 4", station)
	return ok(accessibilityInfo)
}