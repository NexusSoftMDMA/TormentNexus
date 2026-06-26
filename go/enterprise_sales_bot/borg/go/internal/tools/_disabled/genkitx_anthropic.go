package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleAnthropicChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	model, _ :=getString(args, "model")
	prompt, _ :=getString(args, "prompt")
	if apiKey == "" {
		return err("missing api_key")
	}
	if prompt == "" {
		return err("missing prompt")
	}
	if model == "" {
		model = "claude-3-haiku-20240307"
	}
	body := map[string]interface{}{
		"model": model,
		"max_tokens": 1024,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	payload, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(payload))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
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
	return success("Anthropic chat completed")
}

func HandleAnthropicCount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("missing text")
	}
	return success("counted " + string(rune(len(text))) + " characters")
}// touch 1781132126
