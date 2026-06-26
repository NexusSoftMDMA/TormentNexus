package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type classifyResponse struct {
	Class      string  `json:"class"`
	Confidence float64 `json:"confidence"`
}

type listResponse struct {
	Classes []string `json:"classes"`
}

func HandleClassifyAudio(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "audio_url")
	if url == "" {
		return err("audio_url is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.hear-world.ai/v1/classify?url="+url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result classifyResponse
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(fmt.Sprintf("Class: %s (confidence: %.4f)", result.Class, result.Confidence))
}

func HandleListClasses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.hear-world.ai/v1/classes", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result listResponse
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(fmt.Sprintf("Available classes (%d): %v", len(result.Classes), result.Classes))
}