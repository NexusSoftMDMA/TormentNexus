package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type PredictiveLLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type PredictiveLLMRequest struct {
	Model       string                 `json:"model"`
	Messages    []PredictiveLLMMessage `json:"messages"`
	MaxTokens   int                    `json:"max_tokens"`
	Temperature float32                `json:"temperature"`
}

type PredictiveLLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// PredictTools calls FreeLLM local proxy to select the best tools for the current query/goal context.
func PredictTools(ctx context.Context, query string, availableTools []string) ([]string, error) {
	if len(availableTools) == 0 {
		return []string{}, nil
	}

	// Limit to top 50 tools to keep prompt token size small and execution fast
	if len(availableTools) > 50 {
		availableTools = availableTools[:50]
	}

	prompt := fmt.Sprintf(`You are an expert developer assistant orchestrator.
Given the developer's objective/query and the list of available tools, predict and select the top 3-5 tools that are most relevant to solve the objective.

Objective/Query: %s
Available Tools: %s

Respond ONLY with a JSON array of strings containing the selected tool names, e.g. ["read_file", "grep_search"]. Do not include any markdown format blocks or extra text.`, query, strings.Join(availableTools, ", "))

	reqPayload := PredictiveLLMRequest{
		Model: "free-llm",
		Messages: []PredictiveLLMMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   150,
		Temperature: 0.1,
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:4000/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("local LLM proxy call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("local LLM proxy returned status %d", resp.StatusCode)
	}

	var response PredictiveLLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		return []string{}, nil
	}

	rawContent := response.Choices[0].Message.Content
	// Clean up any potential markdown wraps (e.g. ```json ... ```)
	rawContent = strings.TrimPrefix(rawContent, "```json")
	rawContent = strings.TrimPrefix(rawContent, "```")
	rawContent = strings.TrimSuffix(rawContent, "```")
	rawContent = strings.TrimSpace(rawContent)

	var predictedTools []string
	if err := json.Unmarshal([]byte(rawContent), &predictedTools); err != nil {
		// Fallback to simple comma-separation parsing or return empty if LLM response is not formatted correctly
		return []string{}, nil
	}

	return predictedTools, nil
}
