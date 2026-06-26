package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetWedding(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "wedding_id")
	if id == "" {
		return err("wedding_id is required")
}

	resp, e := http.DefaultClient.Get("https://api.zola.com/weddings/" + id)
	if e != nil {
		return err("failed to fetch wedding: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok("wedding info retrieved")
}

func HandleGetGuest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	guestID, _ :=getString(args, "guest_id")
	if guestID == "" {
		return err("guest_id is required")
}

	resp, e := http.DefaultClient.Get("https://api.zola.com/guests/" + guestID)
	if e != nil {
		return err("failed to fetch guest: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok("guest info retrieved")
}