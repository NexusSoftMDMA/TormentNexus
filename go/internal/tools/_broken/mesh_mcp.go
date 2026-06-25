package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// HandleMeshStatus checks the status of a mesh endpoint
func HandleMeshStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		urlStr = "https://httpbin.org/status/200"
	}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch status: %v", fetchErr))
}

	defer resp.Body.Close()

	status := fmt.Sprintf("Status: %d", resp.StatusCode)
	return ok(status)
}

// HandleMeshQuery performs a GET request with query parameters
func HandleMeshQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "url")
	if baseURL == "" {
		return err("url parameter is required")
}

	params := url.Values{}
	if val := getString(args, "key"); val != "" {
		params.Add("key", val)

	if val := getString(args, "value"); val != "" {
		params.Add("value", val)

	if len(params) > 0 {
		if strings.Contains(baseURL, "?") {
			baseURL += "&" + params.Encode()
		} else {
			baseURL += "?" + params.Encode()

	}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to execute request: %v", fetchErr))
}

	defer resp.Body.Close()

	result := fmt.Sprintf("Success: %d bytes read", resp.ContentLength)
	return ok(result)
}

}
}
}

// HandleMeshEcho echoes back the provided text with a timestamp
func HandleMeshEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		text = "No text provided"
	}

	timestamp := time.Now().Format(time.RFC3339)
	response := fmt.Sprintf("[%s] %s", timestamp, text)
	return ok(response)
}

// HandleMeshParseJSON parses a JSON string and returns a specific field
func HandleMeshParseJSON(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	jsonStr, _ :=getString(args, "json")
	if jsonStr == "" {
		return err("json parameter is required")
}

	fieldName, _ :=getString(args, "field")
	if fieldName == "" {
		return err("field parameter is required")
}

	var data map[string]interface{}
	parseErr := json.Unmarshal([]byte(jsonStr), &data)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", parseErr))
}

	val, exists := data[fieldName]
	if !exists {
		return err(fmt.Sprintf("field '%s' not found in JSON", fieldName))
}

	valStr := fmt.Sprintf("%v", val)
	return ok(valStr)
}

// HandleMeshSleep simulates a delay and returns completion message
func HandleMeshSleep(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	duration, _ :=getInt(args, "seconds")
	if duration <= 0 {
		duration = 1
	}

	select {
	case <-time.After(time.Duration(duration) * time.Second):
		return ok(fmt.Sprintf("Slept for %d seconds", duration))
}
	case <-ctx.Done():
		return err("context cancelled during sleep")

}

// HandleMeshHeaders fetches headers from a URL
func HandleMeshHeaders(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		urlStr = "https://httpbin.org/headers"
	}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch headers: %v", fetchErr))
}

	defer resp.Body.Close()

	var headerList []string
	for k, v := range resp.Header {
		headerList = append(headerList, fmt.Sprintf("%s: %s", k, strings.Join(v, ", ")))

	result := strings.Join(headerList, "\n")
	return ok(result)
}
}