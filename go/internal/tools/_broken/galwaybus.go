package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

func HandleGetRoutes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := os.Getenv("GALWAYBUS_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.galwaybus.example.com"
	}

	reqURL := baseURL + "/routes"
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch routes: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
}

	var routes []map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&routes); parseErr != nil {
		return err(fmt.Sprintf("failed to parse routes: %v", parseErr))
}

	routesJSON, marshalErr := json.MarshalIndent(routes, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal routes: %v", marshalErr))
}

	return ok(string(routesJSON))
}

func HandleGetStops(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	routeID, _ :=getString(args, "route_id")
	if routeID == "" {
		return err("missing required parameter: route_id")
}

	baseURL := os.Getenv("GALWAYBUS_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.galwaybus.example.com"
	}

	u, urlErr := url.Parse(baseURL)
	if urlErr != nil {
		return err(fmt.Sprintf("invalid base URL: %v", urlErr))
}

	u.Path = "/stops"
	query := u.Query()
	query.Set("route_id", routeID)
	u.RawQuery = query.Encode()

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch stops: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
}

	var stops []map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&stops); parseErr != nil {
		return err(fmt.Sprintf("failed to parse stops: %v", parseErr))
}

	stopsJSON, marshalErr := json.MarshalIndent(stops, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal stops: %v", marshalErr))
}

	return ok(string(stopsJSON))
}

func HandleGetArrivals(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	stopID, _ :=getString(args, "stop_id")
	if stopID == "" {
		return err("missing required parameter: stop_id")
}

	baseURL := os.Getenv("GALWAYBUS_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.galwaybus.example.com"
	}

	u, urlErr := url.Parse(baseURL)
	if urlErr != nil {
		return err(fmt.Sprintf("invalid base URL: %v", urlErr))
}

	u.Path = "/arrivals"
	query := u.Query()
	query.Set("stop_id", stopID)
	u.RawQuery = query.Encode()

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch arrivals: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
}

	var arrivals []map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&arrivals); parseErr != nil {
		return err(fmt.Sprintf("failed to parse arrivals: %v", parseErr))
}

	arrivalsJSON, marshalErr := json.MarshalIndent(arrivals, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal arrivals: %v", marshalErr))
}

	return ok(string(arrivalsJSON))
}