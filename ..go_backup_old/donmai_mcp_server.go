package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetFleetStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fleetId, _ :=getString(args, "fleet_id")
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://api.donmai.com"
	}
	url := fmt.Sprintf("%s/v1/fleets/%s", baseURL, fleetId)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get fleet status: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Fleet status: %v", result))
}

func HandleListFleets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://api.donmai.com"
	}
	url := fmt.Sprintf("%s/v1/fleets", baseURL)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to list fleets: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var fleets []interface{}
	if e := json.Unmarshal(body, &fleets); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Fleets: %v", fleets))
}