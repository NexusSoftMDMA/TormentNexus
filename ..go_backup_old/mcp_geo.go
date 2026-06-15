package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleGeocode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	addr, _ :=getString(args, "address")
	if addr == "" {
		return err("address is required")
}

	u := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1", url.QueryEscape(addr))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("User-Agent", "MCP-Geo-Server")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	var results []map[string]interface{}
	if e := json.Unmarshal(body, &results); e != nil {
		return err("parse failed: " + e.Error())
}

	if len(results) == 0 {
		return err("no results found")
}

	r := results[0]
	lat, found := r["lat"].(string)
	if !found {
		return err("lat not found")
}

	lon, found := r["lon"].(string)
	if !found {
		return err("lon not found")
}

	return ok(fmt.Sprintf("Latitude: %s, Longitude: %s", lat, lon))
}

func HandleReverseGeocode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat, _ :=getString(args, "lat")
	lon, _ :=getString(args, "lon")
	if lat == "" || lon == "" {
		return err("lat and lon are required")
}

	u := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?lat=%s&lon=%s&format=json", url.QueryEscape(lat), url.QueryEscape(lon))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("User-Agent", "MCP-Geo-Server")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	displayName, found := result["display_name"].(string)
	if !found {
		return err("display_name not found")
}

	return ok(displayName)
}