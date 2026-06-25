package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// ToolResponse, ok(), err("error"), getString(), getInt(), getBool(), TextContent は他の場所で定義されていると仮定

// HandleGetServiceStatus retrieves the health status of the Unla MCP Gateway service
func HandleGetServiceStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "http://localhost:5234"
	}
	healthURL := baseURL + "/health"
	client := http.DefaultClient
	resp, apiErr := client.Get(healthURL)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()

	var health map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&health)
	if parseErr != nil {
		return err(parseErr.Error())
	}

	status, found := health["status"].(string)
	if !found {
		return ok(""), err("status field is missing or not a string")
}

	return ok(map[string]interface{}{"status": status})
}