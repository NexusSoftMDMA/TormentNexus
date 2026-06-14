package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetCurrentWeather(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://data.weather.gov.hk/weatherAPI/opendata/weather.php?dataType=rhrread&lang=en"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch current weather: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("Current weather: %v", data))
}

func HandleGetWeatherForecast(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://data.weather.gov.hk/weatherAPI/opendata/weather.php?dataType=flw&lang=en"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch forecast: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("Weather forecast: %v", data))
}