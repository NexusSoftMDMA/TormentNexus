package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetWeather_mcp_weather(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	if city == "" {
		return err("city is required")
}

	url := fmt.Sprintf("https://wttr.in/%s?format=%%C+%%t", city)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch weather: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}