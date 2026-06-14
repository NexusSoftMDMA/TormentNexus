package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetRelayStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	teamID, _ :=getString(args, "team_id")
	if teamID == "" {
		return err("team_id is required")
}

	url := fmt.Sprintf("https://api.example.com/relay/%s", teamID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch relay status: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	status, found := result["status"].(string)
	if !found {
		return err("status not found in response")
}

	return ok(status)
}

func HandleSetRelay(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	teamID, _ :=getString(args, "team_id")
	enable, _ :=getBool(args, "enable")
	if teamID == "" {
		return err("team_id is required")
}

	url := fmt.Sprintf("https://api.example.com/relay/%s?enable=%t", teamID, enable)
	resp, e := http.DefaultClient.Post(url, "application/json", nil)
	if e != nil {
		return err("failed to set relay: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	return ok("relay updated")
}