package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
)

func HandleGeocode_gis_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	url := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1", address)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to geocode: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if e := json.Unmarshal(body, &results); e != nil {
		return err("failed to parse JSON")
}

	if len(results) == 0 {
		return err("address not found")
}

	return ok(fmt.Sprintf("Latitude: %s, Longitude: %s", results[0].Lat, results[0].Lon))
}

func HandleDistance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat1Str, _ :=getString(args, "lat1")
	lon1Str, _ :=getString(args, "lon1")
	lat2Str, _ :=getString(args, "lat2")
	lon2Str, _ :=getString(args, "lon2")
	lat1, e := strconv.ParseFloat(lat1Str, 64)
	if e != nil {
		return err("invalid lat1")
}

	lon1, e := strconv.ParseFloat(lon1Str, 64)
	if e != nil {
		return err("invalid lon1")
}

	lat2, e := strconv.ParseFloat(lat2Str, 64)
	if e != nil {
		return err("invalid lat2")
}

	lon2, e := strconv.ParseFloat(lon2Str, 64)
	if e != nil {
		return err("invalid lon2")
}

	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c
	return ok(fmt.Sprintf("Distance: %.2f km", distance))
}