package tools

import (
	"context"
	"encoding/json"
)

func HandleAqaraGetDevices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	deviceType, _ :=getString(args, "device_type")
	if deviceType == "" {
		deviceType = "all"
	}
	devices := []map[string]string{
		{"id": "device_1", "name": "Living Room Light", "type": "light"},
		{"id": "device_2", "name": "Kitchen Sensor", "type": "sensor"},
	}
	data, e := json.Marshal(map[string]interface{}{
		"devices": devices,
		"filter":  deviceType,
	})
	if e != nil {
		return err("failed to marshal devices")
}

	return ok(string(data))
}