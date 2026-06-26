package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleRunModel_replicate_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	if model == "" {
		return err("model is required")
}

	inputStr, _ :=getString(args, "input")
	var input map[string]interface{}
	if inputStr != "" {
		if e := json.Unmarshal([]byte(inputStr), &input); e != nil {
			return err("invalid input JSON: " + e.Error())

	}
	body := map[string]interface{}{
		"version": model,
		"input":   input,
	}
	b, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.replicate.com/v1/predictions", bytes.NewReader(b))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	token := os.Getenv("REPLICATE_API_TOKEN")
	if token == "" {
		return err("REPLICATE_API_TOKEN not set")
}

	req.Header.Set("Authorization", "Token "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return err(fmt.Sprintf("API error (status %d): %v", resp.StatusCode, errResp))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("Prediction created: %v", result))
}
}