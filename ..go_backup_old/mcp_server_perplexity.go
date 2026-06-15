package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandlePerplexitySearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiKey := os.Getenv("PERPLEXITY_API_KEY")
	if apiKey == "" {
		return err("PERPLEXITY_API_KEY not set")
}

	body := map[string]interface{}{
		"model": "sonar-pro",
		"messages": []map[string]string{
			{"role": "system", "content": "Be precise and concise."},
			{"role": "user", "content": query},
		},
		"max_tokens": 512,
	}
	payload, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.perplexity.ai/chat/completions", io.NopCloser(nil))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(nil)
	req.Body = io.NopCloser(nil)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API error: status " + resp.Status)
}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	if len(result.Choices) == 0 {
		return err("no response from Perplexity")
}

	return success(result.Choices[0].Message.Content)
}