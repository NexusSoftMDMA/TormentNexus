package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetDeviceInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	deviceID, _ :=getString(args, "device_id")
	if deviceID == "" {
		deviceID = "default"
	}
	url := fmt.Sprintf("http://localhost:8080/device/%s", deviceID)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to get device info")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON response")
}

	return ok(fmt.Sprintf("Device info: %v", result))
}

func HandleListApps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "http://localhost:8080/apps"
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to list apps")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var apps []string
	if e := json.Unmarshal(body, &apps); e != nil {
		return err("invalid JSON response")
}

	return ok(fmt.Sprintf("Apps: %v", apps))
}