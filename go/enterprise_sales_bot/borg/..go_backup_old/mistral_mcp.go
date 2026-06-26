package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		return err("MISTRAL_API_KEY not set")
}

	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	model, _ :=getString(args, "model")
	if model == "" {
		model = "mistral-tiny"
	}
	url := "https://api.mistral.ai/v1/chat/completions"
	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	bodyBytes, e := json.Marshal(reqBody)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	if len(result.Choices) == 0 {
		return err("no choices returned")
}

	return ok(result.Choices[0].Message.Content)
}