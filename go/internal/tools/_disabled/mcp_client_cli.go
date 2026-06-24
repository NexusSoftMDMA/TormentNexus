package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

func HandleRunPrompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	model, _ :=getString(args, "model")
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return err("OPENAI_API_KEY not set")
}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	bodyBytes, e := json.Marshal(reqBody)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(bodyBytes)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("API returned status " + resp.Status)
}

	var respData struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&respData); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if len(respData.Choices) == 0 {
		return err("no choices returned")
}

	return ok(respData.Choices[0].Message.Content)
}