package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetDevices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.thinqconnect.com/devices")
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("devices: %v", data))
}

func HandleGetDevice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "deviceId")
	if id == "" {
		return err("deviceId is required")
}

	url := fmt.Sprintf("https://api.thinqconnect.com/devices/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("device %s: %v", id, data))
}