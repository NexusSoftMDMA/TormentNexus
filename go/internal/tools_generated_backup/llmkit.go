package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("'prompt' is required")
}

	model, _ :=getString(args, "model")
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	maxTokens, _ :=getInt(args, "max_tokens")
	if maxTokens == 0 {
		maxTokens = 100
	}
	url, _ :=getString(args, "url")
	if url == "" {
		url = "https://api.llmkit.com/v1/chat/completions"
	}
	reqBody, _ := json.Marshal(map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": maxTokens,
	})
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(reqBody)))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	if len(result.Choices) == 0 {
		return err("no choices in response")
}

	return ok(result.Choices[0].Message.Content)
}