package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleGetWeather_weather_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	if city == "" {
		return err("city argument is required")
}

	url := "https://wttr.in/" + city + "?format=%C+%t"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch weather: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("weather API returned status " + http.StatusText(resp.StatusCode))
}

	return ok(string(body))
}