package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleConfigureMCPServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	baseUrl, _ :=getString(args, "baseUrl")
	sourceId, _ :=getString(args, "sourceId")
	name, _ :=getString(args, "name")

	if apiKey == "" || baseUrl == "" {
		return err("Missing required fields: apiKey, baseUrl")
	}

	payload := map[string]string{
		"apiKey":   apiKey,
		"baseUrl":  baseUrl,
		"sourceId": sourceId,
		"name":     name,
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("Failed to marshal payload: " + e.Error())
	}

	req, e := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(baseUrl, "/")+"/api/v1/mcp/configure", strings.NewReader(string(body)))
	if e != nil {
		return err("Failed to create request: " + e.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return err("Server returned error: " + resp.Status)
	}

	return success("MCP server configured successfully")
}