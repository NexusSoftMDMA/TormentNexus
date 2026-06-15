package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetIP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.ipify.org?format=json")
	if e != nil {
		return err("failed to get IP")
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON")
	}
	ip, found := result["ip"].(string)
	if !found {
		return err("IP not found")
	}
	return success("Your IP: " + ip)
}

func HandleGetWeather(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	url := "https://wttr.in/" + city + "?format=%C+%t"
	if city == "" {
		url = "https://wttr.in?format=%C+%t"
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get weather")
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
	}
	return success("Weather: " + string(body))
}