package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetAvailableSlots(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	date, _ :=getString(args, "date")
	if date == "" {
		return err("date is required")
}

	resp := map[string]interface{}{
		"available": []string{"09:00", "10:00", "11:00"},
	}
	body, e := json.Marshal(resp)
	if e != nil {
		return err("failed to marshal slots")
}

	return ok(string(body))
}

func HandleScheduleMeeting(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	slot, _ :=getString(args, "slot")
	title, _ :=getString(args, "title")
	if slot == "" || title == "" {
		return err("slot and title are required")
}

	return success("meeting scheduled at " + slot)
}