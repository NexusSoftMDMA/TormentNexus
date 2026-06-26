package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleStageRoom(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	roomType, _ :=getString(args, "room_type")
	style, _ :=getString(args, "style")
	if roomType == "" || style == "" {
		return err("room_type and style are required")
	}
	req := map[string]string{"room": roomType, "style": style}
	body, e := json.Marshal(req)
	if e != nil {
		return err("failed to marshal request")
	}
	resp, e := http.DefaultClient.Post("https://api.immostage.example/stage", "application/json", strings.NewReader(string(body)))
	if e != nil {
		return err("API request failed")
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
	}
	description, found := result["description"].(string)
	if !found {
		return err("unexpected response format")
	}
	return ok(description)
}

func HandleGenerateDescription(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	rooms, _ :=getString(args, "rooms")
	features, _ :=getString(args, "features")
	if rooms == "" {
		return err("rooms is required")
	}
	req := map[string]string{"rooms": rooms, "features": features}
	body, e := json.Marshal(req)
	if e != nil {
		return err("failed to marshal request")
	}
	resp, e := http.DefaultClient.Post("https://api.immostage.example/describe", "application/json", strings.NewReader(string(body)))
	if e != nil {
		return err("API request failed")
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
	}
	description, found := result["description"].(string)
	if !found {
		return err("unexpected response format")
	}
	return ok(description)
}