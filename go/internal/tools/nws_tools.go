//go:build ignore
// +build ignore

package tools

/**
 * @file nws_tools.go
 * @module go/internal/tools
 *
 * WHAT: National Weather Service (NWS) API tools for parity coverage.
 * Provides forecasts, alerts, observations, stations, and office discussions.
 *
 * WHY: Parity — the NWS API is free, requires no auth key, and gives
 * agents weather-aware capabilities matching Claude Code / Codex.
 *
 * API docs: https://www.weather.gov/documentation/services-web-api
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const nwsBase = "https://api.weather.gov"

// nwsResponse wraps common NWS response fields.
type nwsResponse struct {
	Properties json.RawMessage `json:"properties"`
	Features   []nwsFeature   `json:"features"`
	Status     string         `json:"status,omitempty"`
	Title      string         `json:"title,omitempty"`
	Detail     string         `json:"detail,omitempty"`
}

type nwsFeature struct {
	Properties json.RawMessage `json:"properties"`
	ID         string          `json:"id"`
}

// nwsGet fetches an NWS API endpoint and returns the raw response.
func nwsGet(ctx context.Context, path string, args map[string]interface{}) (*nwsResponse, error) {
	timeoutSec := getInt(args, "timeout")
	if timeoutSec <= 0 {
		timeoutSec = 15
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	apiURL := nwsBase + path
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return nil, fmt.Errorf("error creating request: %w", reqErr)
	}
	req.Header.Set("User-Agent", "TormentNexus/NWS-Parity/1.0 (github.com/tormentnexushq/tormentnexus-go)")
	req.Header.Set("Accept", "application/json")

	resp, doErr := http.DefaultClient.Do(req)
	if doErr != nil {
		return nil, fmt.Errorf("request failed: %w", doErr)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("error reading response: %w", readErr)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NWS API returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result nwsResponse
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return nil, fmt.Errorf("error parsing response: %w", parseErr)
	}

	return &result, nil
}

// formatProperties formats a raw properties JSON into indented key-value text.
func formatProperties(props json.RawMessage) string {
	var data map[string]interface{}
	if err := json.Unmarshal(props, &data); err != nil {
		return string(props)
	}
	var sb strings.Builder
	for k, v := range data {
		sb.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
	}
	return strings.TrimSpace(sb.String())
}

// nwsGetGridPoints gets the grid point metadata for a lat/lon.
func nwsGetGridPoints(ctx context.Context, lat, lon string, args map[string]interface{}) (string, error) {
	path := fmt.Sprintf("/points/%.4f,%.4f", parseFloat(lat), parseFloat(lon))
	result, nwsErr := nwsGet(ctx, path, args)
	if nwsErr != nil {
		return "", nwsErr
	}

	var props struct {
		GridID  string `json:"gridId"`
		GridX   int    `json:"gridX"`
		GridY   int    `json:"gridY"`
	}
	if parseErr := json.Unmarshal(result.Properties, &props); parseErr != nil {
		return "", fmt.Errorf("error parsing grid point data: %w", parseErr)
	}
	if props.GridID == "" {
		return "", fmt.Errorf("no grid data found for this location")
	}

	return fmt.Sprintf("/gridpoints/%s/%d,%d", props.GridID, props.GridX, props.GridY), nil
}

// HandleNWSGetForecast gets the forecast for a point (lat,lon).
func HandleNWSGetForecast(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat, _ := getString(args, "latitude", "lat")
	lon, _ := getString(args, "longitude", "lon")
	if lat == "" || lon == "" {
		return err("latitude and longitude are required")
	}

	gridPath, gridErr := nwsGetGridPoints(ctx, lat, lon, args)
	if gridErr != nil {
		return err(gridErr.Error())
	}

	result, nwsErr := nwsGet(ctx, gridPath+"/forecast", args)
	if nwsErr != nil {
		return err(nwsErr.Error())
	}

	props := formatProperties(result.Properties)
	return ok(fmt.Sprintf("Forecast:\n%s", props))
}

// HandleNWSGetHourlyForecast gets an hourly forecast for a point.
func HandleNWSGetHourlyForecast(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat, _ := getString(args, "latitude", "lat")
	lon, _ := getString(args, "longitude", "lon")
	if lat == "" || lon == "" {
		return err("latitude and longitude are required")
	}

	gridPath, gridErr := nwsGetGridPoints(ctx, lat, lon, args)
	if gridErr != nil {
		return err(gridErr.Error())
	}

	result, nwsErr := nwsGet(ctx, gridPath+"/forecast/hourly", args)
	if nwsErr != nil {
		return err(nwsErr.Error())
	}

	props := formatProperties(result.Properties)
	return ok(fmt.Sprintf("Hourly Forecast:\n%s", props))
}

// HandleNWSSearchAlerts searches for active weather alerts.
func HandleNWSSearchAlerts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	queryVals := url.Values{}
	if area, ok := getString(args, "area", "region"); ok {
		queryVals.Set("area", area)
	}
	if event, ok := getString(args, "event", "type"); ok {
		queryVals.Set("event", event)
	}
	if code, ok := getString(args, "code", "code"); ok {
		queryVals.Set("code", code)
	}
	if state, ok := getString(args, "state", "state"); ok {
		queryVals.Set("state", state)
	}
	if limit := getInt(args, "limit", "max_results"); limit > 0 {
		queryVals.Set("limit", fmt.Sprintf("%d", limit))
	} else {
		queryVals.Set("limit", "10")
	}

	path := "/alerts/active"
	if qs := queryVals.Encode(); qs != "" {
		path += "?" + qs
	}

	result, nwsErr := nwsGet(ctx, path, args)
	if nwsErr != nil {
		return err(nwsErr.Error())
	}

	if len(result.Features) == 0 {
		return ok("No active alerts found.")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d active alerts:\n\n", len(result.Features)))
	for i, f := range result.Features {
		if i >= 10 {
			sb.WriteString(fmt.Sprintf("... and %d more alerts\n", len(result.Features)-10))
			break
		}
		sb.WriteString(fmt.Sprintf("--- Alert %d ---\n%s\n\n", i+1, formatProperties(f.Properties)))
	}

	return ok(strings.TrimSpace(sb.String()))
}

// HandleNWSGetObservations gets recent observations for a station.
func HandleNWSGetObservations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	station, _ := getString(args, "station", "station_id")
	if station == "" {
		return err("station is required (e.g., KLAX)")
	}

	limit := getInt(args, "limit", "max_results")
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	path := fmt.Sprintf("/stations/%s/observations?limit=%d", url.PathEscape(strings.ToUpper(station)), limit)
	result, nwsErr := nwsGet(ctx, path, args)
	if nwsErr != nil {
		return err(nwsErr.Error())
	}

	if len(result.Features) == 0 {
		return ok("No observations found.")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Recent observations for station %s:\n\n", station))
	for i, f := range result.Features {
		if i >= limit {
			break
		}
		sb.WriteString(fmt.Sprintf("--- Observation %d ---\n%s\n\n", i+1, formatProperties(f.Properties)))
	}

	return ok(strings.TrimSpace(sb.String()))
}

// HandleNWSFindStations finds stations near a point.
func HandleNWSFindStations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat, _ := getString(args, "latitude", "lat")
	lon, _ := getString(args, "longitude", "lon")
	if lat == "" || lon == "" {
		return err("latitude and longitude are required")
	}

	path := fmt.Sprintf("/points/%.4f,%.4f/stations", parseFloat(lat), parseFloat(lon))
	result, nwsErr := nwsGet(ctx, path, args)
	if nwsErr != nil {
		return err(nwsErr.Error())
	}

	if len(result.Features) == 0 {
		return ok("No stations found near this location.")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Weather stations near (%.4f, %.4f):\n\n", parseFloat(lat), parseFloat(lon)))
	for i, f := range result.Features {
		if i >= 10 {
			break
		}
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, formatProperties(f.Properties)))
		sb.WriteString("\n")
	}

	return ok(strings.TrimSpace(sb.String()))
}

// HandleNWSListAlertTypes lists all valid NWS alert types.
func HandleNWSListAlertTypes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	result, nwsErr := nwsGet(ctx, "/alerts/types", args)
	if nwsErr != nil {
		return err(nwsErr.Error())
	}

	typeList := struct {
		EventTypes []struct {
			Event string `json:"event"`
		} `json:"eventTypes"`
	}{}

	if parseErr := json.Unmarshal(result.Properties, &typeList); parseErr != nil {
		// Fallback: just show raw data
		return ok(formatProperties(result.Properties))
	}

	var sb strings.Builder
	sb.WriteString("NWS Alert Types:\n\n")
	for _, et := range typeList.EventTypes {
		sb.WriteString(fmt.Sprintf("  - %s\n", et.Event))
	}

	return ok(strings.TrimSpace(sb.String()))
}

// HandleNWSGetOfficeDiscussion gets the forecast discussion from an office.
func HandleNWSGetOfficeDiscussion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	office, _ := getString(args, "office", "office_id")
	if office == "" {
		return err("office is required (e.g., LOX)")
	}

	path := fmt.Sprintf("/products/types/AFD/locations/%s", url.PathEscape(strings.ToUpper(office)))
	result, nwsErr := nwsGet(ctx, path, args)
	if nwsErr != nil {
		return err(nwsErr.Error())
	}

	if len(result.Features) == 0 {
		return ok(fmt.Sprintf("No discussion found for office %s.", office))
	}

	// Show the latest discussion
	latest := result.Features[0]
	return ok(fmt.Sprintf("Latest discussion for office %s:\n\n%s", office, formatProperties(latest.Properties)))
}

// HandleNWSGetZoneForecast gets a forecast for a specific zone.
func HandleNWSGetZoneForecast(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	zone, _ := getString(args, "zone", "zone_id")
	if zone == "" {
		return err("zone is required (e.g., CAZ511)")
	}

	path := fmt.Sprintf("/zones/forecast/%s/forecast", url.PathEscape(strings.ToUpper(zone)))
	result, nwsErr := nwsGet(ctx, path, args)
	if nwsErr != nil {
		return err(nwsErr.Error())
	}

	props := formatProperties(result.Properties)
	return ok(fmt.Sprintf("Zone forecast for %s:\n%s", zone, props))
}

// parseFloat converts a latitude/longitude string to float64.
func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
