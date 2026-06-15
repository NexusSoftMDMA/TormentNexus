package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetFlight(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	flightNumber, _ :=getString(args, "flightNumber")
	if flightNumber == "" {
		return err("flightNumber is required")
}

	url := fmt.Sprintf("https://api.variflight.com/flight/%s", flightNumber)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch flight: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON response")
}

	return ok(result)
}

func HandleSearchFlights(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	from, _ :=getString(args, "from")
	to, _ :=getString(args, "to")
	if from == "" || to == "" {
		return err("both 'from' and 'to' are required")
}

	url := fmt.Sprintf("https://api.variflight.com/search?from=%s&to=%s", from, to)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("search request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read search results")
}

	var data []interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON response")
}

	return success(fmt.Sprintf("Found %d flights", len(data)))
}