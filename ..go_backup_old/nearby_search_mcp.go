package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

func HandleNearbySearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	latStr, _ :=getString(args, "lat")
	lngStr, _ :=getString(args, "lng")
	radius, _ :=getInt(args, "radius")
	keyword, _ :=getString(args, "keyword")
	apiKey, _ :=getString(args, "api_key")
	if latStr == "" || lngStr == "" || apiKey == "" {
		return err("Missing required parameters lat, lng, or api_key")
}

	lat, e := strconv.ParseFloat(latStr, 64)
	if e != nil {
		return err("Invalid lat")
}

	lng, e := strconv.ParseFloat(lngStr, 64)
	if e != nil {
		return err("Invalid lng")
}

	if radius <= 0 {
		radius = 1000
	}
	u, _ := url.Parse("https://maps.googleapis.com/maps/api/place/nearbysearch/json")
	q := u.Query()
	q.Set("location", fmt.Sprintf("%f,%f", lat, lng))
	q.Set("radius", strconv.Itoa(radius))
	q.Set("key", apiKey)
	if keyword != "" {
		q.Set("keyword", keyword)

	u.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Read response failed: " + e.Error())
}

	var result struct {
		Results []map[string]interface{} `json:"results"`
		Status  string                   `json:"status"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("JSON parse failed: " + e.Error())
}

	if result.Status != "OK" {
		return err("API error: " + result.Status)
}

	out, _ := json.Marshal(result.Results)
	return ok(fmt.Sprintf("Nearby places: %s", string(out)))
}
}