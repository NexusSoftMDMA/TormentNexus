package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const galwayBusBaseURL = "https://data.galwaybus.ie/api"

var galwayBusClient = http.DefaultClient

// BusStop represents a physical bus stop location
type BusStop struct {
	StopID      string  `json:"stopId"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Description string  `json:"description"`
}

// BusRoute represents a bus route line
type BusRoute struct {
	RouteID   string `json:"routeId"`
	LongName  string `json:"longName"`
	ShortName string `json:"shortName"`
}

// BusArrival represents real-time arrival information
type BusArrival struct {
	Route       string `json:"route"`
	Destination string `json:"destination"`
	Duetime     string `json:"duetime"`
	Timestamp   string `json:"timestamp"`
}

// StopInfoResponse wraps the response for stop info
type StopInfoResponse struct {
	StopID    string       `json:"stopId"`
	Timestamp string       `json:"timestamp"`
	Results   []BusArrival `json:"results"`
}

// HandleGetStops retrieves the list of all Galway Bus stops
func HandleGetStops(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	reqURL := fmt.Sprintf("%s/stops", galwayBusBaseURL)

	resp, fetchErr := galwayBusClient.Get(reqURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status %d", resp.StatusCode))
}

	var stops []BusStop
	decodeErr := json.NewDecoder(resp.Body).Decode(&stops)
	if decodeErr != nil {
		return err(decodeErr.Error())
}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d bus stops:\n\n", len(stops)))
	for _, s := range stops {
		sb.WriteString(fmt.Sprintf("- [%s] %s (Lat: %.4f, Lon: %.4f)\n", s.StopID, s.Description, s.Latitude, s.Longitude))

	return ok(sb.String())
}

}

// HandleGetRoutes retrieves the list of all Galway Bus routes
func HandleGetRoutes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	reqURL := fmt.Sprintf("%s/routes", galwayBusBaseURL)

	resp, fetchErr := galwayBusClient.Get(reqURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status %d", resp.StatusCode))
}

	var routes []BusRoute
	decodeErr := json.NewDecoder(resp.Body).Decode(&routes)
	if decodeErr != nil {
		return err(decodeErr.Error())
}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d bus routes:\n\n", len(routes)))
	for _, r := range routes {
		sb.WriteString(fmt.Sprintf("- [%s] %s\n", r.RouteID, r.LongName))

	return ok(sb.String())
}

}

// HandleGetStopInfo retrieves real-time arrival information for a specific stop
func HandleGetStopInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	stopID, _ :=getString(args, "stopId")
	if stopID == "" {
		return err("argument 'stopId' is required")
}

	params := url.Values{}
	params.Add("stopId", stopID)
	reqURL := fmt.Sprintf("%s/stops?%s", galwayBusBaseURL, params.Encode())

	resp, fetchErr := galwayBusClient.Get(reqURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status %d", resp.StatusCode))
}

	var stopInfo StopInfoResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&stopInfo)
	if decodeErr != nil {
		return err(decodeErr.Error())
}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Arrivals for stop [%s] at %s:\n\n", stopInfo.StopID, stopInfo.Timestamp))
	if len(stopInfo.Results) == 0 {
		sb.WriteString("No arrivals found.")
	} else {
		for _, arrival := range stopInfo.Results {
			sb.WriteString(fmt.Sprintf("- Route %s to %s, due in %s\n", arrival.Route, arrival.Destination, arrival.Duetime))

	}

	return ok(sb.String())
}
}