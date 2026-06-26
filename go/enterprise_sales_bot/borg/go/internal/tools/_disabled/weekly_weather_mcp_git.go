package tools

import "context"

func HandleGetWeather(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	if city == "" {
		return err("city parameter is required")
}

	return success("Current weather in " + city + ": 72°F, Sunny")
}