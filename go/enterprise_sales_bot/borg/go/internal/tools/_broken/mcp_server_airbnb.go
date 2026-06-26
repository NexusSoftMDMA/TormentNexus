package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func HandleSearchAirbnb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	location, _ :=getString(args, "location")
	checkIn, _ :=getString(args, "checkIn")
	checkOut, _ :=getString(args, "checkOut")
	adults, _ :=getInt(args, "adults")
	if adults == 0 {
		adults = 1
	}

	// Simulate API call using a free placeholder
	url := fmt.Sprintf("https://api.mocki.io/v1/abcdef?location=%s&checkIn=%s&checkOut=%s&adults=%d", location, checkIn, checkOut, adults)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch listings: " + e.Error())
}

	defer resp.Body.Close()

	var result struct {
		Listings []map[string]interface{} `json:"listings"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response")
}

	if len(result.Listings) == 0 {
		return ok("No Airbnb listings found for the given criteria.")
}

	msg := fmt.Sprintf("Found %d Airbnb listings:\n", len(result.Listings))
	for _, l := range result.Listings {
		msg += fmt.Sprintf("- %s (price: %v)\n", l["title"], l["price"])

	return ok(msg)
}

}

func HandleCurrentTime_mcp_server_airbnb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	return ok("Current time: " + now)
}