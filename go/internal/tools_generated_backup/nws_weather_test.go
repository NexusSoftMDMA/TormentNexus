package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandleNWSTools(t *testing.T) {
	// Start mock NWS API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/geo+json")

		path := r.URL.Path
		switch {
		case strings.HasPrefix(path, "/points/47.6062,-122.3321"):
			w.Write([]byte(`{
				"properties": {
					"gridId": "SEW",
					"gridX": 125,
					"gridY": 125,
					"forecast": "http://` + r.Host + `/gridpoints/SEW/125,125/forecast",
					"forecastHourly": "http://` + r.Host + `/gridpoints/SEW/125,125/forecast/hourly",
					"observationStations": "http://` + r.Host + `/gridpoints/SEW/125,125/stations",
					"relativeLocation": {
						"properties": {
							"city": "Seattle",
							"state": "WA"
						}
					},
					"timeZone": "America/Los_Angeles",
					"forecastZone": "https://api.weather.gov/zones/forecast/WAZ315",
					"county": "https://api.weather.gov/zones/county/WAC033"
				}
			}`))
		case strings.HasPrefix(path, "/gridpoints/SEW/125,125/forecast/hourly"):
			w.Write([]byte(`{
				"properties": {
					"generatedAt": "2026-06-05T00:00:00Z",
					"periods": [
						{
							"number": 1,
							"name": "Hourly Forecast",
							"startTime": "2026-06-05T00:00:00Z",
							"endTime": "2026-06-05T01:00:00Z",
							"temperature": 68,
							"temperatureUnit": "F",
							"windSpeed": "10 mph",
							"windDirection": "NW",
							"shortForecast": "Mostly Sunny",
							"detailedForecast": "Mostly Sunny with a high near 68."
						}
					]
				}
			}`))
		case strings.HasPrefix(path, "/gridpoints/SEW/125,125/forecast"):
			w.Write([]byte(`{
				"properties": {
					"generatedAt": "2026-06-05T00:00:00Z",
					"periods": [
						{
							"number": 1,
							"name": "Today",
							"startTime": "2026-06-05T08:00:00Z",
							"endTime": "2026-06-05T20:00:00Z",
							"temperature": 70,
							"temperatureUnit": "F",
							"windSpeed": "5 mph",
							"windDirection": "W",
							"shortForecast": "Sunny",
							"detailedForecast": "Sunny with a high near 70."
						}
					]
				}
			}`))
		case path == "/alerts/active":
			w.Write([]byte(`{
				"features": [
					{
						"properties": {
							"id": "alert1",
							"event": "Special Weather Statement",
							"severity": "minor",
							"description": "A minor weather update for Seattle"
						}
					}
				]
			}`))
		case path == "/gridpoints/SEW/125,125/stations":
			w.Write([]byte(`{
				"features": [
					{
						"geometry": {
							"coordinates": [-122.3321, 47.6062]
						},
						"properties": {
							"stationIdentifier": "KBFI",
							"name": "Boeing Field",
							"timeZone": "America/Los_Angeles",
							"forecast": "https://api.weather.gov/zones/forecast/WAZ315",
							"county": "https://api.weather.gov/zones/county/WAC033"
						}
					}
				]
			}`))
		case path == "/stations/KBFI":
			w.Write([]byte(`{
				"properties": {
					"name": "Boeing Field",
					"timeZone": "America/Los_Angeles"
				}
			}`))
		case path == "/stations/KBFI/observations/latest":
			w.Write([]byte(`{
				"properties": {
					"timestamp": "2026-06-05T01:00:00Z",
					"textDescription": "Partly Cloudy",
					"temperature": {"value": 20.0, "unitCode": "wmoUnit:degC"},
					"windSpeed": {"value": 5.0, "unitCode": "wmoUnit:m_s-1"}
				}
			}`))
		case path == "/alerts/types":
			w.Write([]byte(`{
				"eventTypes": ["Tornado Warning", "Special Weather Statement"]
			}`))
		case path == "/products/types/AFD/locations/SEW":
			w.Write([]byte(`{
				"@graph": [
					{
						"id": "latest-afd-id"
					}
				]
			}`))
		case path == "/products/latest-afd-id":
			w.Write([]byte(`{
				"issuanceTime": "2026-06-05T00:30:00Z",
				"issuingOffice": "KSEW",
				"productCode": "AFD",
				"productName": "Area Forecast Discussion",
				"productText": "Meteorological reasoning text here",
				"wmoCollectiveId": "FXUS66"
			}`))
		case path == "/zones/forecast/WAZ315/forecast":
			w.Write([]byte(`{
				"properties": {
					"updated": "2026-06-05T00:00:00Z",
					"periods": [
						{
							"number": 1,
							"name": "Today",
							"detailedForecast": "Sunny. Highs in the lower 70s."
						}
					]
				}
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	os.Setenv("NWS_API_URL", server.URL)
	defer os.Unsetenv("NWS_API_URL")

	// Update base URL function mapping
	nwsBaseURL = server.URL

	ctx := context.Background()

	// Test 1: nws_get_forecast
	resp, err := HandleNWSGetForecast(ctx, map[string]interface{}{
		"latitude":  47.6062,
		"longitude": -122.3321,
	})
	if err != nil {
		t.Fatalf("HandleNWSGetForecast failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "Seattle") {
		t.Errorf("Expected city Seattle, got: %s", resp.Content[0].Text)
	}

	// Test 1b: nws_get_forecast (hourly)
	resp, err = HandleNWSGetForecast(ctx, map[string]interface{}{
		"latitude":  47.6062,
		"longitude": -122.3321,
		"hourly":    true,
	})
	if err != nil {
		t.Fatalf("HandleNWSGetForecast hourly failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "Hourly Forecast") {
		t.Errorf("Expected hourly forecast, got: %s", resp.Content[0].Text)
	}

	// Test 2: nws_search_alerts
	resp, err = HandleNWSSearchAlerts(ctx, map[string]interface{}{
		"area": "WA",
	})
	if err != nil {
		t.Fatalf("HandleNWSSearchAlerts failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "Special Weather Statement") {
		t.Errorf("Expected alert in WA, got: %s", resp.Content[0].Text)
	}

	// Test 2b: nws_search_alerts with event filter
	resp, err = HandleNWSSearchAlerts(ctx, map[string]interface{}{
		"area":  "WA",
		"event": []string{"Tornado", "Special"},
	})
	if err != nil {
		t.Fatalf("HandleNWSSearchAlerts with event filter failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "Special Weather Statement") {
		t.Errorf("Expected filtered alert, got: %s", resp.Content[0].Text)
	}

	// Test 3: nws_get_observations (with coords)
	resp, err = HandleNWSGetObservations(ctx, map[string]interface{}{
		"latitude":  47.6062,
		"longitude": -122.3321,
	})
	if err != nil {
		t.Fatalf("HandleNWSGetObservations failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "Partly Cloudy") {
		t.Errorf("Expected observation description, got: %s", resp.Content[0].Text)
	}

	// Test 3b: nws_get_observations (with stationId)
	resp, err = HandleNWSGetObservations(ctx, map[string]interface{}{
		"station_id": "KBFI",
	})
	if err != nil {
		t.Fatalf("HandleNWSGetObservations with station ID failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "Partly Cloudy") {
		t.Errorf("Expected observation description, got: %s", resp.Content[0].Text)
	}

	// Test 4: nws_find_stations
	resp, err = HandleNWSFindStations(ctx, map[string]interface{}{
		"latitude":  47.6062,
		"longitude": -122.3321,
		"limit":     5,
	})
	if err != nil {
		t.Fatalf("HandleNWSFindStations failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "Boeing Field") {
		t.Errorf("Expected station name, got: %s", resp.Content[0].Text)
	}

	// Test 5: nws_list_alert_types
	resp, err = HandleNWSListAlertTypes(ctx, nil)
	if err != nil {
		t.Fatalf("HandleNWSListAlertTypes failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "Tornado Warning") {
		t.Errorf("Expected alert types, got: %s", resp.Content[0].Text)
	}

	// Test 6: nws_get_office_discussion
	resp, err = HandleNWSGetOfficeDiscussion(ctx, map[string]interface{}{
		"office":       "SEW",
		"product_type": "AFD",
	})
	if err != nil {
		t.Fatalf("HandleNWSGetOfficeDiscussion failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "Meteorological reasoning") {
		t.Errorf("Expected AFD text, got: %s", resp.Content[0].Text)
	}

	// Test 7: nws_get_zone_forecast
	resp, err = HandleNWSGetZoneForecast(ctx, map[string]interface{}{
		"zone_id": "WAZ315",
	})
	if err != nil {
		t.Fatalf("HandleNWSGetZoneForecast failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "lower 70s") {
		t.Errorf("Expected zone forecast narrative, got: %s", resp.Content[0].Text)
	}
}
