package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// HandleGetClimateSystemState gets the current climate system state
func HandleGetClimateSystemState(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Get Home Assistant base URL from environment
	baseURL := os.Getenv("HOMEASSISTANT_URL")
	if baseURL == "" {
		baseURL = "http://homeassistant.local:8123"
	}

	// Get API token from environment
	apiToken := os.Getenv("HOMEASSISTANT_TOKEN")
	if apiToken == "" {
		return err("HOMEASSISTANT_TOKEN environment variable not set")
}

	// Construct URL for climate system state sensor
	reqURL := fmt.Sprintf("%s/api/states/input_text.climate_system_state", baseURL)

	// Create request
	req, e := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := http.DefaultClient
	resp, e := client.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to get climate system state: %v", e))
}

	defer resp.Body.Close()

	// Read response
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	// Parse response
	var stateData map[string]interface{}
	if e := json.Unmarshal(body, &stateData); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	// Extract state value
	state, found := stateData["state"].(string)
	if !found {
		state = "unknown"
	}

	return ok(fmt.Sprintf("Climate system state: %s", state))
}

// HandleControlBoiler controls the boiler state
func HandleControlBoiler(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Get parameters
	action, _ :=getString(args, "action")
	if action == "" {
		return err("action parameter is required")
}

	baseURL := os.Getenv("HOMEASSISTANT_URL")
	if baseURL == "" {
		baseURL = "http://homeassistant.local:8123"
	}

	apiToken := os.Getenv("HOMEASSISTANT_TOKEN")
	if apiToken == "" {
		return err("HOMEASSISTANT_TOKEN environment variable not set")
}

	// Construct URL for boiler service
	reqURL := fmt.Sprintf("%s/api/services/homeassistant/turn_%s", baseURL, action)

	// Create request body
	requestBody := map[string]interface{}{
		"entity_id": "climate.boiler",
	}

	jsonBody, e := json.Marshal(requestBody)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal request body: %v", e))
}

	// Create request
	req, e := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(string(jsonBody)))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := http.DefaultClient
	resp, e := client.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to control boiler: %v", e))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("boiler control failed: %s", string(body)))
}

	return ok(fmt.Sprintf("Boiler %s command sent successfully", action))
}

// HandleGetSensorReading gets a sensor reading
func HandleGetSensorReading(ctx context.Context, args map[string