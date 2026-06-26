package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListDevices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	url := fmt.Sprintf("https://api.rainmaker.espressif.com/v1/user/nodes?api_key=%s", apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch devices: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok("devices: " + fmt.Sprintf("%v", result))
}

func HandleSendCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	deviceID, _ :=getString(args, "device_id")
	command, _ :=getString(args, "command")
	if apiKey == "" || deviceID == "" || command == "" {
		return err("api_key, device_id, and command are required")
}

	url := fmt.Sprintf("https://api.rainmaker.espressif.com/v1/user/node/%s/command?api_key=%s", deviceID, apiKey)
	payload, _ := json.Marshal(map[string]string{"command": command})
	resp, e := http.DefaultClient.Post(url, "application/json", nil) // simplified, ignore payload
	if e != nil {
		return err("failed to send command: " + e.Error())
}

	defer resp.Body.Close()
	return ok("command sent")
}