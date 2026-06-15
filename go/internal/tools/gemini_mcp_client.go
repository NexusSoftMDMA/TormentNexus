//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGeminiChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	model, _ :=getString(args, "model")
	if model == "" {
		model = "gemini-1.5-flash"
	}
	if prompt == "" {
		return err("missing prompt argument")
	}
	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": prompt}}},
		},
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, getString(args, "api_key")), bytes.NewReader(body))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response")
	}
	return success(fmt.Sprintf("Gemini response: %v", result))
}

func HandleGeminiCountTokens(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("missing text argument")
	}
	return success(fmt.Sprintf("Token counting for '%s' requires specific endpoint implementation", text))
}
