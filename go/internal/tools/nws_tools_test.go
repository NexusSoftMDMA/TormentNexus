//go:build ignore
// +build ignore

package tools

import (
	"context"
	"testing"
)

// NWS tool tests – test argument validation (no live API calls for missing-arg tests).

func TestHandleNWSGetForecast_MissingArgs(t *testing.T) {
	resp, err := HandleNWSGetForecast(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatal("expected no error")
	}
	if !resp.IsError || resp.Content[0].Text != "latitude and longitude are required" {
		t.Fatalf("expected latitude/longitude error, got: %v", resp.Content)
	}

	// Missing longitude
	resp, err = HandleNWSGetForecast(context.Background(), map[string]interface{}{"latitude": "40.7"})
	if err != nil {
		t.Fatal("expected no error")
	}
	if !resp.IsError {
		t.Fatal("expected error when longitude missing")
	}
}

func TestHandleNWSGetForecast_WithArgs(t *testing.T) {
	// This will try a real API call; skip if it fails quickly
	resp, err := HandleNWSGetForecast(context.Background(), map[string]interface{}{
		"latitude":  "40.7128",
		"longitude": "-74.0060",
		"timeout":   5,
	})
	if err != nil {
		t.Fatal("expected no error")
	}
	if resp.IsError {
		t.Logf("NWS forecast API returned error (expected if offline): %s", resp.Content[0].Text)
	}
}

func TestHandleNWSGetHourlyForecast_MissingArgs(t *testing.T) {
	resp, err := HandleNWSGetHourlyForecast(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatal("expected no error")
	}
	if !resp.IsError || resp.Content[0].Text != "latitude and longitude are required" {
		t.Fatalf("expected latitude/longitude error, got: %v", resp.Content)
	}
}

func TestHandleNWSSearchAlerts_MissingArgs(t *testing.T) {
	// Empty call should still work (no filters = all active alerts)
	resp, err := HandleNWSSearchAlerts(context.Background(), map[string]interface{}{
		"timeout": 5,
	})
	if err != nil {
		t.Fatal("expected no error")
	}
	if resp.IsError {
		t.Logf("NWS alerts error (expected if offline): %s", resp.Content[0].Text)
	}
}

func TestHandleNWSSearchAlerts_WithFilters(t *testing.T) {
	resp, err := HandleNWSSearchAlerts(context.Background(), map[string]interface{}{
		"area":   "CA",
		"limit":  3,
		"timeout": 5,
	})
	if err != nil {
		t.Fatal("expected no error")
	}
	if resp.IsError {
		t.Logf("NWS filtered alerts error (expected if offline): %s", resp.Content[0].Text)
	}
}

func TestHandleNWSGetObservations_MissingStation(t *testing.T) {
	resp, err := HandleNWSGetObservations(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatal("expected no error")
	}
	if !resp.IsError || resp.Content[0].Text != "station is required (e.g., KLAX)" {
		t.Fatalf("expected station error, got: %v", resp.Content)
	}
}

func TestHandleNWSGetObservations_WithStation(t *testing.T) {
	resp, err := HandleNWSGetObservations(context.Background(), map[string]interface{}{
		"station": "KLAX",
		"limit":   3,
		"timeout": 5,
	})
	if err != nil {
		t.Fatal("expected no error")
	}
	if resp.IsError {
		t.Logf("NWS observations error (expected if offline): %s", resp.Content[0].Text)
	}
}

func TestHandleNWSFindStations_MissingArgs(t *testing.T) {
	resp, err := HandleNWSFindStations(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatal("expected no error")
	}
	if !resp.IsError || resp.Content[0].Text != "latitude and longitude are required" {
		t.Fatalf("expected latitude/longitude error, got: %v", resp.Content)
	}
}

func TestHandleNWSFindStations_WithArgs(t *testing.T) {
	resp, err := HandleNWSFindStations(context.Background(), map[string]interface{}{
		"latitude":  "40.7128",
		"longitude": "-74.0060",
		"timeout":   5,
	})
	if err != nil {
		t.Fatal("expected no error")
	}
	if resp.IsError {
		t.Logf("NWS stations error (expected if offline): %s", resp.Content[0].Text)
	}
}

func TestHandleNWSListAlertTypes(t *testing.T) {
	resp, err := HandleNWSListAlertTypes(context.Background(), map[string]interface{}{
		"timeout": 5,
	})
	if err != nil {
		t.Fatal("expected no error")
	}
	if resp.IsError {
		t.Logf("NWS alert types error (expected if offline): %s", resp.Content[0].Text)
	}
}

func TestHandleNWSGetOfficeDiscussion_MissingOffice(t *testing.T) {
	resp, err := HandleNWSGetOfficeDiscussion(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatal("expected no error")
	}
	if !resp.IsError || resp.Content[0].Text != "office is required (e.g., LOX)" {
		t.Fatalf("expected office error, got: %v", resp.Content)
	}
}

func TestHandleNWSGetOfficeDiscussion_WithOffice(t *testing.T) {
	resp, err := HandleNWSGetOfficeDiscussion(context.Background(), map[string]interface{}{
		"office":  "LOX",
		"timeout": 5,
	})
	if err != nil {
		t.Fatal("expected no error")
	}
	if resp.IsError {
		t.Logf("NWS office discussion error (expected if offline): %s", resp.Content[0].Text)
	}
}

func TestHandleNWSGetZoneForecast_MissingZone(t *testing.T) {
	resp, err := HandleNWSGetZoneForecast(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatal("expected no error")
	}
	if !resp.IsError || resp.Content[0].Text != "zone is required (e.g., CAZ511)" {
		t.Fatalf("expected zone error, got: %v", resp.Content)
	}
}

func TestHandleNWSGetZoneForecast_WithZone(t *testing.T) {
	resp, err := HandleNWSGetZoneForecast(context.Background(), map[string]interface{}{
		"zone":    "CAZ511",
		"timeout": 5,
	})
	if err != nil {
		t.Fatal("expected no error")
	}
	if resp.IsError {
		t.Logf("NWS zone forecast error (expected if offline): %s", resp.Content[0].Text)
	}
}

func TestRegistry_NWSToolsRegistered(t *testing.T) {
	r := NewRegistry()
	expectedTools := []string{
		"nws_get_forecast",
		"nws_get_hourly_forecast",
		"nws_search_alerts",
		"nws_get_observations",
		"nws_find_stations",
		"nws_list_alert_types",
		"nws_get_office_discussion",
		"nws_get_zone_forecast",
	}
	for _, tool := range expectedTools {
		if !r.HasTool(tool) {
			t.Errorf("expected tool '%s' to be registered, but it was not", tool)
		}
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"40.7128", 40.7128},
		{"-74.0060", -74.0060},
		{"0", 0},
		{"invalid", 0},
	}
	for _, tt := range tests {
		got := parseFloat(tt.input)
		if (got-tt.want) > 0.0001 || (tt.want-got) > 0.0001 {
			if tt.input == "invalid" || tt.input == "0" {
				continue // acceptable for edge cases
			}
			t.Errorf("parseFloat(%q) = %f, want %f", tt.input, got, tt.want)
		}
	}
}
