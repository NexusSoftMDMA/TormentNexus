package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetWeather(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	if city == "" {
		return err("city argument is required")
}

	url := fmt.Sprintf("https://wttr.in/%s?format=j1", city)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
}

	return success(fmt.Sprintf("Weather for %s retrieved", city))
}

func HandleListCities(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Available cities: London, New York, Tokyo")
}