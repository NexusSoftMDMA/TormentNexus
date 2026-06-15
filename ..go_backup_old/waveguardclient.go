package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleWaveguardListDevices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	devices := []map[string]interface{}{
		{"id": "device-1", "name": "Sensor A"},
		{"id": "device-2", "name": "Sensor B"},
	}
	data, e := json.Marshal(devices)
	if e != nil {
		return err("failed to marshal devices")
	}
	return ok(fmt.Sprintf(`{"devices": %s}`, string(data)))
}

func HandleWaveguardGetDeviceInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	deviceID, _ :=getString(args, "deviceId")
	if deviceID == "" {
		return err("missing deviceId parameter")
	}
	info := map[string]string{"id": deviceID, "status": "active", "signal": "good"}
	data, e := json.Marshal(info)
	if e != nil {
		return err("failed to marshal info")
	}
	return ok(fmt.Sprintf(`{"device": %s}`, string(data)))
}