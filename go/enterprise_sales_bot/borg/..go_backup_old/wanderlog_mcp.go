package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

type Trip struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Destinations []string `json:"destinations"`
}

func HandleListTrips(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	trips := []Trip{
		{ID: "1", Title: "Paris Getaway", Destinations: []string{"Eiffel Tower", "Louvre"}},
		{ID: "2", Title: "Tokyo Adventure", Destinations: []string{"Shibuya", "Mount Fuji"}},
	}
	if len(trips) > limit {
		trips = trips[:limit]
	}
	data, e := json.Marshal(trips)
	if e != nil {
		return err("failed to marshal trips")
}

	return ok(string(data))
}

func HandleGetTrip(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "trip_id")
	if id == "" {
		return err("trip_id is required")
}

	trips := map[string]Trip{
		"1": {ID: "1", Title: "Paris Getaway", Destinations: []string{"Eiffel Tower", "Louvre"}},
		"2": {ID: "2", Title: "Tokyo Adventure", Destinations: []string{"Shibuya", "Mount Fuji"}},
	}
	trip, found := trips[id]
	if !found {
		return err(fmt.Sprintf("trip not found: %s", id))
}

	data, e := json.Marshal(trip)
	if e != nil {
		return err("failed to marshal trip")
}

	return ok(string(data))
}