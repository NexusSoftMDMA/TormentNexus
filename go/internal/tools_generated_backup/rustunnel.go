package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListTunnels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/tunnels", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return success(fmt.Sprintf("%v", data))
}

func HandleCreateTunnel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	payload := map[string]interface{}{
		"name":       getString(args, "name"),
		"local_port": getInt(args, "local_port"),
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", baseURL+"/tunnels", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return err("unexpected status: " + resp.Status)
}

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(respBody, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success(fmt.Sprintf("Tunnel created: %v", data))
}