package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchFlights(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	origin, _ :=getString(args, "origin")
	destination, _ :=getString(args, "destination")
	departureDate, _ :=getString(args, "departureDate")
	apiKey, _ :=getString(args, "apiKey")
	if origin == "" || destination == "" || departureDate == "" || apiKey == "" {
		return err("Missing required parameters: origin, destination, departureDate, apiKey")
}

	baseURL := "https://test.api.amadeus.com/v2/shopping/flight-offers"
	reqURL := fmt.Sprintf("%s?origin=%s&destination=%s&departureDate=%s", baseURL, url.QueryEscape(origin), url.QueryEscape(destination), url.QueryEscape(departureDate))
	req, e := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if e != nil {
		return err(fmt.Sprintf("Failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("Request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("Failed to decode response: %v", e))
}

	return ok(fmt.Sprintf("Found %d flight offers", len(result)))
}