package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
)

func HandleGetCoordinates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q := url.QueryEscape(getString(args, "location"))
	resp, e := http.DefaultClient.Get("https://nominatim.openstreetmap.org/search?q=" + q + "&format=json&limit=1")
	if e != nil {
		return err("failed to call geocoding API: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	json.Unmarshal(body, &results)
	if len(results) == 0 {
		return err("no results found")
}

	return ok(fmt.Sprintf("lat: %s, lon: %s", results[0].Lat, results[0].Lon))
}

func HandleGetDistance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	toRad := func(s string) float64 {
		f, _ := parseFloat(s)
		return f * math.Pi / 180
	}
	lat1 := toRad(getString(args, "lat1"))
	lon1 := toRad(getString(args, "lon1"))
	lat2 := toRad(getString(args, "lat2"))
	lon2 := toRad(getString(args, "lon2"))
	dLat := lat2 - lat1
	dLon := lon2 - lon1
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	km := 6371 * c
	return ok(fmt.Sprintf("distance: %.2f km", km))
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, e := fmt.Sscanf(strings.TrimSpace(s), "%f", &f)
	return f, e
}