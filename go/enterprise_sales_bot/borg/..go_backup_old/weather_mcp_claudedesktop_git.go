package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetWeather(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	if city == "" {
		return err("city is required")
}

	resp, e := http.DefaultClient.Get("https://wttr.in/" + city + "?format=%t")
	if e != nil {
		return err("failed to fetch weather: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	temp := string(body)
	if temp == "" {
		return err("no weather data")
}

	return ok("Temperature in " + city + ": " + temp)
}