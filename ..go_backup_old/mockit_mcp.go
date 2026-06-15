package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleGenerateMockup(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("API key is required")
}

	body := map[string]interface{}{
		"model":     "claude-opus-4-7",
		"max_tokens": 2048,
		"messages":  []map[string]string{{"role": "user", "content": prompt}},
	}
	b, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(b))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode: " + e.Error())
}

	content, found := result["content"].([]interface{})
	if !found || len(content) == 0 {
		return err("no content")
}

	text, found := content[0].(map[string]interface{})["text"].(string)
	if !found {
		return err("no text")
}

	return success("Mockup generated: " + text)
}