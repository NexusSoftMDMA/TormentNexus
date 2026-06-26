package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetWeather_mcp_weather_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	if city == "" {
		return err("city is required")
}

	url := fmt.Sprintf("https://wttr.in/%s?format=j1", city)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch weather: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		CurrentCondition []struct {
			TempC          string `json:"temp_C"`
			WeatherDesc    []struct {
				Value string `json:"value"`
			} `json:"weatherDesc"`
			Humidity       string `json:"humidity"`
			WindspeedKmph  string `json:"windspeedKmph"`
		} `json:"current_condition"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse weather data: " + e.Error())
}

	if len(result.CurrentCondition) == 0 {
		return err("no weather data for " + city)
}

	c := result.CurrentCondition[0]
	desc := ""
	if len(c.WeatherDesc) > 0 {
		desc = c.WeatherDesc[0].Value
	}
	msg := fmt.Sprintf("Weather in %s: %s, Temperature: %s°C, Humidity: %s%%, Wind: %s km/h", city, desc, c.TempC, c.Humidity, c.WindspeedKmph)
	return ok(msg)
}