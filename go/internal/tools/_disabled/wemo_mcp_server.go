package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// HandleListDevices returns a list of known Wemo devices.
func HandleListDevices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	devices := []map[string]string{
		{"id": "1", "name": "Living Room Light", "ip": "192.168.1.10"},
		{"id": "2", "name": "Bedroom Switch", "ip": "192.168.1.11"},
	}
	data, e := json.Marshal(devices)
	if e != nil {
		return err("failed to marshal devices")
}

	return ok(string(data))
}