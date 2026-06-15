package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleGeocode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	apiURL := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1", url.QueryEscape(address))
	req, e := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("User-Agent", "ChatspatialMCP/1.0")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	var results []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&results); e != nil {
		return err("decode error")
}

	if len(results) == 0 {
		return err("no results")
}

	lat := results[0]["lat"]
	lon := results[0]["lon"]
	displayName := results[0]["display_name"]
	return ok(fmt.Sprintf("Location: %v, %v - %v", lat, lon, displayName))
}

func HandleReverseGeocode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat, _ :=getString(args, "lat")
	lon, _ :=getString(args, "lon")
	if lat == "" || lon == "" {
		return err("lat and lon are required")
}

	apiURL := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?lat=%s&lon=%s&format=json", url.QueryEscape(lat), url.QueryEscape(lon))
	req, e := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("User-Agent", "ChatspatialMCP/1.0")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error")
}

	displayName, found := result["display_name"].(string)
	if !found {
		return err("no display name")
}

	return ok(fmt.Sprintf("Address: %s", displayName))
}