package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGeocodeForward(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	token := os.Getenv("MAPBOX_ACCESS_TOKEN")
	url := fmt.Sprintf("https://api.mapbox.com/geocoding/v5/mapbox.places/%s.json?access_token=%s", query, token)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var data struct {
		Features []struct {
			Center []float64 `json:"center"`
		} `json:"features"`
	}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(e.Error())
}

	if len(data.Features) == 0 {
		return err("no results")
}

	center := data.Features[0].Center
	result := fmt.Sprintf("Longitude: %f, Latitude: %f", center[0], center[1])
	return ok(result)
}

func HandleGeocodeReverse(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lng, _ :=getString(args, "longitude")
	lat, _ :=getString(args, "latitude")
	if lng == "" || lat == "" {
		return err("longitude and latitude are required")
}

	token := os.Getenv("MAPBOX_ACCESS_TOKEN")
	url := fmt.Sprintf("https://api.mapbox.com/geocoding/v5/mapbox.places/%s,%s.json?access_token=%s", lng, lat, token)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var data struct {
		Features []struct {
			PlaceName string `json:"place_name"`
		} `json:"features"`
	}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(e.Error())
}

	if len(data.Features) == 0 {
		return err("no results")
}

	return ok(data.Features[0].PlaceName)
}// touch 1781132131
