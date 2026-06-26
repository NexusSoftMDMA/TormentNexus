package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetClient_honeybook_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	clientID, _ :=getString(args, "clientId")
	if clientID == "" {
		return err("clientId is required")
}

	url := fmt.Sprintf("https://api.honeybook.com/v1/clients/%s", clientID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer YOUR_API_KEY")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(data)
}

func HandleListClients_honeybook_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.honeybook.com/v1/clients"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer YOUR_API_KEY")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	clients, found := data["clients"].([]interface{})
	if !found {
		return err("unexpected response format")
}

	return success(clients)
}