package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetCurrentWeather(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat, _ :=getString(args, "latitude")
	lon, _ :=getString(args, "longitude")
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&current=temperature_2m,relative_humidity_2m,apparent_temperature,precipitation,weather_code,wind_speed_10m", lat, lon)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode: " + e.Error())
}

	data, e := json.Marshal(result)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	return ok(string(data))
}

func HandleGetForecast(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat, _ :=getString(args, "latitude")
	lon, _ :=getString(args, "longitude")
	days, _ :=getString(args, "forecast_days")
	if days == "" {
		days = "7"
	}
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&daily=temperature_2m_max,temperature_2m_min,precipitation_sum,weather_code&forecast_days=%s", lat, lon, days)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode: " + e.Error())
}

	data, e := json.Marshal(result)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	return ok(string(data))
}