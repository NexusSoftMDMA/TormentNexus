package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleGetDevices_plugwise_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
	}
	url, _ :=getString(args, "base_url") + "/devices?api_key=" + apiKey
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch devices: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	return ok(string(body))
}

func HandleGetDeviceStatus_plugwise_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	deviceID, _ :=getString(args, "device_id")
	if apiKey == "" || deviceID == "" {
		return err("api_key and device_id are required")
	}
	url, _ :=getString(args, "base_url") + "/devices/" + deviceID + "?api_key=" + apiKey
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch device status: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	return ok(string(body))
}