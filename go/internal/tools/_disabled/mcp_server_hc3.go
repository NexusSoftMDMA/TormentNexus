package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListDevices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://hc3.local/api/devices")
	if e != nil {
		return err(fmt.Sprintf("failed to fetch devices: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var devices []map[string]interface{}
	if e := json.Unmarshal(body, &devices); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return success(string(body))
}

func HandleGetDeviceStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("device id is required")
}

	url := fmt.Sprintf("http://hc3.local/api/devices/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch device: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return success(string(body))
}