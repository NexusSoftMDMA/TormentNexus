package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleWeatherForecast(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat, _ :=getString(args, "latitude")
	lon, _ :=getString(args, "longitude")
	if lat == "" || lon == "" {
		return err("latitude and longitude are required")
}

	pointsURL := fmt.Sprintf("https://api.weather.gov/points/%s,%s", lat, lon)
	resp, e := http.DefaultClient.Get(pointsURL)
	if e != nil {
		return err("failed to fetch points: " + e.Error())
}

	defer resp.Body.Close()
	var points struct {
		Properties struct {
			Forecast string `json:"forecast"`
		} `json:"properties"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&points); e != nil {
		return err("failed to decode points: " + e.Error())
}

	if points.Properties.Forecast == "" {
		return err("forecast URL not found")
}

	forecastResp, e := http.DefaultClient.Get(points.Properties.Forecast)
	if e != nil {
		return err("failed to fetch forecast: " + e.Error())
}

	defer forecastResp.Body.Close()
	var forecastData interface{}
	if e := json.NewDecoder(forecastResp.Body).Decode(&forecastData); e != nil {
		return err("failed to decode forecast: " + e.Error())
}

	data, _ := json.MarshalIndent(forecastData, "", "  ")
	return ok(string(data))
}

func HandleSevereWeatherAlerts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	state, _ :=getString(args, "state")
	if state == "" {
		return err("state is required")
}

	alertsURL := fmt.Sprintf("https://api.weather.gov/alerts?area=%s", state)
	resp, e := http.DefaultClient.Get(alertsURL)
	if e != nil {
		return err("failed to fetch alerts: " + e.Error())
}

	defer resp.Body.Close()
	var alertsData interface{}
	if e := json.NewDecoder(resp.Body).Decode(&alertsData); e != nil {
		return err("failed to decode alerts: " + e.Error())
}

	data, _ := json.MarshalIndent(alertsData, "", "  ")
	return ok(string(data))
}