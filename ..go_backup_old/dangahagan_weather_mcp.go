package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetCurrentWeather(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat, _ :=getString(args, "latitude")
	lon, _ :=getString(args, "longitude")
	if lat == "" || lon == "" {
		return err("latitude and longitude are required")
}

	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&current_weather=true&timezone=auto", lat, lon)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get weather data: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success(fmt.Sprintf("Current weather: %v", data["current_weather"]))
}

func GetForecast(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat, _ :=getString(args, "latitude")
	lon, _ :=getString(args, "longitude")
	if lat == "" || lon == "" {
		return err("latitude and longitude are required")
}

	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&daily=temperature_2m_max,temperature_2m_min&timezone=auto&forecast_days=7", lat, lon)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get forecast: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success(fmt.Sprintf("Forecast: %v", data["daily"]))
}// touch 1781132124
