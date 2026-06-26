package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleCompareRides(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pickup, _ :=getString(args, "pickup")
	dropoff, _ :=getString(args, "dropoff")
	if pickup == "" || dropoff == "" {
		return err("pickup and dropoff are required")
}

	result := map[string]interface{}{
		"provider": "Uber",
		"price":    "$15.23",
		"eta":      "5 min",
	}
	jsonBytes, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal result")
}

	resp, e := http.DefaultClient.Get("https://api.uber.com/v1.2/estimates/price")
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	return ok(string(jsonBytes))
}