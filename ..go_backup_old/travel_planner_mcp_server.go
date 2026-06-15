package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetDestination(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	url := "https://api.travelplanner.com/destinations?city=" + city
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch destination")
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response")
}

	return ok(data)
}

func HandlePlanTrip(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	origin, _ :=getString(args, "origin")
	dest, _ :=getString(args, "destination")
	url := "https://api.travelplanner.com/trip?from=" + origin + "&to=" + dest
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to plan trip")
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response")
}

	return ok(data)
}