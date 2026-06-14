package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleIpfind(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ip, _ :=getString(args, "ip")
	if ip == "" {
		return err("ip parameter is required")
}

	resp, e := http.DefaultClient.Get("http://ip-api.com/json/" + ip)
	if e != nil {
		return err("api request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Status  string  `json:"status"`
		City    string  `json:"city"`
		Region  string  `json:"regionName"`
		Country string  `json:"country"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	if result.Status != "success" {
		return err("api returned error for IP: " + ip)
}

	msg := fmt.Sprintf("IP: %s\nCity: %s\nRegion: %s\nCountry: %s\nLat: %f\nLon: %f",
		ip, result.City, result.Region, result.Country, result.Lat, result.Lon)
	return ok(msg)
}