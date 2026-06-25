package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HandleListInstances lists all MCP instances
func HandleListInstances(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit := 20
	if l, found := args["limit"].(float64); found {
		limit = int(l)

	offset := 0
	if o, found := args["offset"].(float64); found {
		offset = int(o)

	instances := []map[string]interface{}{
		{
			"id":         "inst-001",
			"name":       "gateway-primary",
			"type":       "gateway",
			"status":     "running",
			"url":        "https://mcp.example.com",
			"created_at": "2024-01-15T10:30:00Z",
		},
		{
			"id":         "inst-002",
			"name":       "proxy-east",
			"type":       "proxy",
			"status":     "running",
			"url":        "https://proxy-east.example.com/mcp",
			"created_at": "2024-01-16T14:20:00Z",
		},
		{
			"id":         "inst-003",
			"name":       "sse-backend",
			"type":       "sse",
			"status":     "running",
			"url":        "https://sse.example.com/sse",
			"created_at": "2024-01-17T09:15:00Z",
		},
	}

	if offset >= len(instances) {
		instances = []map[string]interface{}{}
	} else if offset+limit < len(instances) {
		instances = instances[offset : offset+limit]
	} else {
		instances = instances[offset:]
	}

	result := map[string]interface{}{
		"instances": instances,
		"total":     3,
		"limit":     limit,
		"offset":    offset,
	}

	data, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(data))
}

}
}

// HandleGetInstance retrieves details of a specific MCP instance
func HandleGetInstance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	instanceID, _ :=getString(args, "instance_id")
	if instanceID == "" {
		return err("instance_id is required")
}

	instance := map[string]interface{}{
		"id":             instanceID,
		"name":           "gateway-" + instanceID,
		"type":           "gateway",
		"status":         "running",
		"url":            "https://mcp.example.com",
		"proxy_protocol": "stdio",
		"config": map[string]interface{}{
			"transport":     "stdio",
			"timeout":       30,
			"retry_count":   3,
			"auto_reconnect": true,
		},
		"created_at": "2024-01-15T10:30:00Z",
		"updated_at": "2024-01-20T16:45:00Z",
	}

	data, jsonErr := json.Marshal(instance)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(data))
}

// HandleCreateInstance creates a new MCP instance
func HandleCreateInstance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	instanceType, _ :=getString(args, "type")
	if instanceType == "" {
		instanceType = "gateway"
	}

	mcpURL, _ :=getString(args, "url")
	proxyProtocol, _ :=getString(args, "proxy_protocol")
	if proxyProtocol == "" {
		proxyProtocol = "stdio"
	}

	instance := map[string]interface{}{
		"id":             fmt.Sprintf("inst-%d", time.Now().UnixNano()),
		"name":           name,
		"type":           instanceType,
		"status":         "created",
		"url":            mcpURL,
		"proxy_protocol": proxyProtocol,
		"created_at":     time.Now().Format(time.RFC3339),
	}

	data, jsonErr := json.Marshal(instance)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(data))
}

// HandleDeleteInstance deletes an MCP instance
func HandleDeleteInstance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	instanceID, _ :=getString(args, "instance_id")
	if instanceID == "" {
		return err("instance_id is required")
}

	result := map[string]interface{}{
		"id":        instanceID,
		"deleted":   true,
		"deleted_at": time.Now().Format(time.RFC3339),
	}

	data, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(data))
}

// HandleHealthCheck checks the health of the MCP gateway
func HandleHealthCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	gatewayURL, _ :=getString(args, "url")
	if gatewayURL == "" {
		gatewayURL = "https://mcp.example.com"
	}

	client := http.DefaultClient

	healthURL := gatewayURL
	if !strings.HasSuffix(gatewayURL, "/health") && !strings.HasSuffix(gatewayURL, "/") {
		healthURL = gatewayURL + "/health"
	}

	resp, fetchErr := client.Get(healthURL)
	if fetchErr != nil {
		result := map[string]interface{}{
			"status":    "unhealthy",
			"url":       gatewayURL,
			"error":     fetchErr.Error(),
			"checked_at": time.Now().Format(time.RFC3339),
		}
		data, jsonErr := json.Marshal(result)
		if jsonErr != nil {
			return err(jsonErr.Error())
}

		return ok(string(data))
}

	defer resp.Body.Close()

	status := "healthy"
	if resp.StatusCode >= 400 {
		status = "unhealthy"
	}

	result := map[string]interface{}{
		"status":     status,
		"url":        gatewayURL,
		"status_code": resp.StatusCode,
		"checked_at": time.Now().Format(time.RFC3339),
	}

	data, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(data))
}

// HandleGetConfig generates MCP configuration for AI sessions
func HandleGetConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	instanceID, _ :=getString(args, "instance_id")
	gatewayURL, _ :=getString(args, "gateway_url")
	transportType, _ :=getString(args, "transport")

	if gatewayURL == "" {
		gatewayURL = "https://mcp.example.com"
	}

	if transportType == "" {
		if strings.HasSuffix(gatewayURL, "/mcp") {
			transportType = "stdio"
		} else if strings.HasSuffix(gatewayURL, "/sse") {
			transportType = "sse"
		} else {
			transportType = "stdio"
		}
	}

	config := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"mcpcan": map[string]interface{}{
				"url": gatewayURL,
			},
		},
		"instance_id": instanceID,
		"transport":   transportType,
		"proxy_protocol": "stdio",
		"auto_correct": true,
	}

	data, jsonErr := json.Marshal(config)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(data))
}

// HandleNormalizeURL normalizes and validates MCP gateway URLs
func HandleNormalizeURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	inputURL, _ :=getString(args, "url")
	if inputURL == "" {
		return err("url is required")
}

	parsedURL, parseErr := url.Parse(inputURL)
	if parseErr != nil {
		return err("invalid URL: " + parseErr.Error())
}

	normalizedPath := parsedURL.Path
	detectedTransport := "stdio"

	if strings.HasSuffix(normalizedPath, "/mcp") {
		detectedTransport = "stdio"
	} else if strings.HasSuffix(normalizedPath, "/sse") {
		detectedTransport = "sse"
	}

	result := map[string]interface{}{
		"original_url":     inputURL,
		"normalized_url":  parsedURL.String(),
		"scheme":          parsedURL.Scheme,
		"host":            parsedURL.Host,
		"path":            normalizedPath,
		"detected_transport": detectedTransport,
		"is_valid":        true,
	}

	data, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(data))
}