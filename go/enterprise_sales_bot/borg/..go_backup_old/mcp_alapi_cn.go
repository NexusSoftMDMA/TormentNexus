package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"io"
)

func HandleWeather(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	if city == "" {
		return err("city is required")
}

	apiURL := "https://v1.alapi.cn/api/weather?city=" + url.QueryEscape(city)
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	data, found := result["data"].(map[string]interface{})
	if !found {
		return err("no data in response")
}

	weather, found := data["weather"].(string)
	if !found {
		weather = "unknown"
	}
	temp, found := data["temperature"].(string)
	if !found {
		temp = "N/A"
	}
	return success("Weather in " + city + ": " + weather + ", temperature: " + temp)
}