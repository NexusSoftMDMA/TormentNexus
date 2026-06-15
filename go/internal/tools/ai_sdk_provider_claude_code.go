//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleClaudeCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	apiKey := os.Getenv("CLAUDE_API_KEY")
	if apiKey == "" {
		return err("CLAUDE_API_KEY not set")
}

	baseURL := os.Getenv("CLAUDE_API_URL")
	if baseURL == "" {

	}
	body := map[string]interface{}{
		"model":      getString(args, "model"),
		"max_tokens": getInt(args, "max_tokens"),
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		if body["model"].(string) == "" {

		if body["max_tokens"].(int) == 0 {

	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("marshal: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", baseURL+"/messages", bytes.NewReader(jsonBody))
	if e != nil {
		return err("request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("do: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{	if e := json.NewDecoder(res


-reasoner (deepseek)*,
}
}
}
}