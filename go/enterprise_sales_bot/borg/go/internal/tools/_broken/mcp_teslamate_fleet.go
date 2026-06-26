package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListVehicles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	token, _ :=getString(args, "token")
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/vehicles", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var vehicles []map[string]interface{}
	if e = json.Unmarshal(body, &vehicles); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d vehicles", len(vehicles)))
}

}

func HandleGetVehicleData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	vehicleID, _ :=getString(args, "vehicle_id")
	if vehicleID == "" {
		return err("vehicle_id is required")
}

	token, _ :=getString(args, "token")
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/vehicles/"+vehicleID, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e = json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Vehicle data: %s", string(body)))
}
}