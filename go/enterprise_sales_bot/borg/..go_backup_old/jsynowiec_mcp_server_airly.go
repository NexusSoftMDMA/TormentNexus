package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func HandleGetCurrentMeasurements(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("AIRLY_API_KEY")
	if apiKey == "" {
		return err("AIRLY_API_KEY environment variable not set")
}

	latStr, _ :=getString(args, "lat")
	lngStr, _ :=getString(args, "lng")
	if latStr == "" || lngStr == "" {
		return err("lat and lng are required")
}

	lat, e := strconv.ParseFloat(latStr, 64)
	if e != nil {
		return err("invalid lat")
}

	lng, e := strconv.ParseFloat(lngStr, 64)
	if e != nil {
		return err("invalid lng")
}

	url := fmt.Sprintf("https://airapi.airly.eu/v2/measurements/nearest?lat=%f&lng=%f&maxDistanceKM=10", lat, lng)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("apikey", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
}

	current, found := data["current"].(map[string]interface{})
	if !found {
		return err("no current data")
}

	values, found := current["values"].([]interface{})
	if !found {
		return err("no values in current")
}

	output := "Current air quality:\n"
	for _, v := range values {
		item, found := v.(map[string]interface{})
		if !found {
			continue
		}
		name, _ := item["name"].(string)
		val, _ := item["value"].(float64)
		output += fmt.Sprintf("%s: %.2f\n", name, val)

	return ok(output)
}
}