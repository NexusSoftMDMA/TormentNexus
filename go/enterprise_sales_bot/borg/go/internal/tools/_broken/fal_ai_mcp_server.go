package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleRunInference(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	modelID, _ :=getString(args, "model_id")
	input, _ :=getString(args, "input")
	apiKey, _ :=getString(args, "api_key")
	if modelID == "" || input == "" || apiKey == "" {
		return err("model_id, input, and api_key are required")
}

	reqBody, e := json.Marshal(map[string]interface{}{
		"model": modelID,
		"input": input,
	})
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://fal.ai/api/v1/run", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(nil)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	requestID, found := result["request_id"].(string)
	if !found || requestID == "" {
		return err("no request_id in response")
}

	return success(fmt.Sprintf("Inference started. Request ID: %s", requestID))
}

func HandleGetInferenceStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	requestID, _ :=getString(args, "request_id")
	apiKey, _ :=getString(args, "api_key")
	if requestID == "" || apiKey == "" {
		return err("request_id and api_key are required")
}

	url := fmt.Sprintf("https://fal.ai/api/v1/requests/%s/status", requestID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	status, found := result["status"].(string)
	if !found {
		return err("no status in response")
}

	return success(fmt.Sprintf("Status: %s", status))
}