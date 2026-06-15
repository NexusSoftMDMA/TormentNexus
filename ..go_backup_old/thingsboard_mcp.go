package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetDeviceTelemetry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	deviceID, _ :=getString(args, "deviceId")
	if deviceID == "" {
		return err("deviceId is required")
}

	keys, _ :=getString(args, "keys")
	if keys == "" {
		return err("keys is required")
}

	startTs, _ :=getInt(args, "startTs")
	endTs, _ :=getInt(args, "endTs")
	if startTs == 0 || endTs == 0 {
		return err("startTs and endTs are required")
}

	baseURL := os.Getenv("THINGSBOARD_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	token := os.Getenv("THINGSBOARD_TOKEN")
	if token == "" {
		return err("THINGSBOARD_TOKEN not set")
}

	url := fmt.Sprintf("%s/api/plugins/telemetry/DEVICE/%s/values/timeseries?keys=%s&startTs=%d&endTs=%d", baseURL, deviceID, keys, startTs, endTs)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("X-Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Telemetry data: %s", string(body)))
}