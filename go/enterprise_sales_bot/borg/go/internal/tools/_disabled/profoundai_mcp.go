package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
)

func HandleGenerateText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	apiKey := os.Getenv("PROFOUND_API_KEY")
	if apiKey == "" {
		return err("PROFOUND_API_KEY not set")
}

	body, _ := json.Marshal(map[string]string{"prompt": prompt})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.profound.ai/v1/generate", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	text, found := result["text"].(string)
	if !found {
		return err("unexpected response format")
}

	return ok(text)
}