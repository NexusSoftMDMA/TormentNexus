package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetNightlifeVenues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	url := fmt.Sprintf("https://api.nightlife.com/venues?city=%s", city)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch venues: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Venues in %s: %v", city, data))
}

func HandleGetNightlifeEvents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	venueType, _ :=getString(args, "type")
	url := fmt.Sprintf("https://api.nightlife.com/events?city=%s&type=%s", city, venueType)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch events: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Events in %s (%s): %v", city, venueType, data))
}