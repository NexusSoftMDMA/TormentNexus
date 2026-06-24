package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListTrips(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("HYPERTRACK_API_KEY")
	if apiKey == "" {
		return err("missing HYPERTRACK_API_KEY")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.hypertrack.com/v1/trips", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to call API: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
	}
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %s", string(body)))
	}
	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
	}
	return ok(fmt.Sprintf("Trips: %s", string(body)))
}

func HandleGetTrip(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tripID, _ :=getString(args, "trip_id")
	if tripID == "" {
		return err("missing trip_id argument")
	}
	apiKey := os.Getenv("HYPERTRACK_API_KEY")
	if apiKey == "" {
		return err("missing HYPERTRACK_API_KEY")
	}
	url := fmt.Sprintf("https://api.hypertrack.com/v1/trips/%s", tripID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to call API: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
	}
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %s", string(body)))
	}
	return ok(fmt.Sprintf("Trip: %s", string(body)))
}