package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListServers_jurislm_hetzner_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "api_token")
	if token == "" {
		return err("api_token is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.hetzner.cloud/v1/servers", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	return ok(fmt.Sprintf("Servers: %s", string(body)))
}

func HandleCreateServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "api_token")
	name, _ :=getString(args, "name")
	serverType, _ :=getString(args, "server_type")
	image, _ :=getString(args, "image")
	if token == "" || name == "" || serverType == "" || image == "" {
		return err("api_token, name, server_type, image are required")
}

	payload := map[string]interface{}{
		"name":        name,
		"server_type": serverType,
		"image":       image,
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("marshal payload: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.hetzner.cloud/v1/servers", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response: %v", e))
}

	if resp.StatusCode != 201 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
}

	return success(fmt.Sprintf("Created server: %s", string(respBody)))
}