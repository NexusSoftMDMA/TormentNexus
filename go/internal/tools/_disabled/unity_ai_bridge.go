package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetScene(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sceneName, _ :=getString(args, "sceneName")
	if sceneName == "" {
		sceneName = "default"
	}
	resp, e := http.DefaultClient.Get("http://localhost:8080/scene/" + sceneName)
	if e != nil {
		return err("failed to fetch scene: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("Scene '%s' data: %v", sceneName, data))
}

func HandleSendCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	payload, _ := json.Marshal(map[string]string{"command": cmd})
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/command", bytes.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send command: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(result.Message)
}