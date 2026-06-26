package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetWeather_gpt_computer_assistant(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	if city == "" {
		return err("City is required")
}

	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&appid=YOUR_API_KEY", city)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()

	var data map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err(e.Error())
}

	return ok(fmt.Sprintf("Weather in %s: %v", city, data["weather"]))
}

func HandleGetTime_gpt_computer_assistant(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	timezone, _ :=getString(args, "timezone")
	if timezone == "" {
		timezone = "UTC"
	}

	url := fmt.Sprintf("http://worldtimeapi.org/api/timezone/%s", timezone)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()

	var data map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err(e.Error())
}

	return ok(fmt.Sprintf("Current time in %s: %v", timezone, data["datetime"]))
}// touch 1781132127
