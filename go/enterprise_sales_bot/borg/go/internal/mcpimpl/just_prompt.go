package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleJustPrompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
	}
	if model == "" {
		model = "gpt-4o"
	}
	body, e := json.Marshal(map[string]interface{}{
		"model":  model,
		"messages": []map[string]string{{"role": "user", "content": prompt}},
	})
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(body)))
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
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed")
	}
	return success("Prompt sent to " + model)
}

func HandleListModels_just_prompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Available models: gpt-4o, claude-3-5-sonnet, gemini-1.5-pro, llama-3-70b, deepseek-coder, ollama")
}