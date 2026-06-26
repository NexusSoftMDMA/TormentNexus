package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetCurrentTrack(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	addr, _ :=getString(args, "address")
	if addr == "" {
		addr = "http://127.0.0.1:9000"
	}
	resp, e := http.DefaultClient.Get(fmt.Sprintf("%s/api/live/track/current", addr))
	if e != nil {
		return err("failed to fetch track: " + e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
	}
	track, found := result["name"]
	if !found {
		return ok("No track currently selected")
	}
	return ok(fmt.Sprintf("Current track: %v", track))
}

func HandleTogglePlay(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	addr, _ :=getString(args, "address")
	if addr == "" {
		addr = "http://127.0.0.1:9000"
	}
	// Use a simple POST to toggle play state (requires body? we'll send an empty json)
	req, e := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/live/play/toggle", addr), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to toggle play: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("toggle returned status %d", resp.StatusCode))
	}
	return ok("Play state toggled")
}