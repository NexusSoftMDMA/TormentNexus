package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListDevices_matter_controller_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := "http://localhost:8080"
	resp, e := http.DefaultClient.Get(base + "/devices")
	if e != nil {
		return err("failed to fetch devices: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Devices: %v", result))
}

func HandleSetDeviceState(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	deviceID, _ :=getString(args, "device_id")
	state, _ :=getString(args, "state")
	if deviceID == "" {
		return err("device_id is required")
}

	if state == "" {
		return err("state is required")
}

	base := "http://localhost:8080"
	url := fmt.Sprintf("%s/devices/%s/state", base, deviceID)
	payload := map[string]string{"state": state}
	jsonData, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(jsonData))
	if e != nil {
		return err("failed to send request: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("request failed with status: " + resp.Status)
}

	return success("Device state updated")
}