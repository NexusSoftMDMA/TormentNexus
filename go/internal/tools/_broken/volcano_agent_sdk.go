package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HandleCallLLM calls an LLM provider with a given prompt and returns the response
func HandleCallLLM(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	provider, _ :=getString(args, "provider")
	if provider == "" {
		return err("provider is required")
}

	model, _ :=getString(args, "model")
	if model == "" {
		return err("model is required")
}

	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	// Route to appropriate provider
	switch strings.ToLower(provider) {
	case "openai":
		return callOpenAI(apiKey, model, prompt)
}
	case "anthropic":
		return callAnthropic(apiKey, model, prompt)
}
	default:
		return err(fmt.Sprintf("unsupported provider: %s", provider))

}

func callOpenAI(apiKey, model, prompt string) (ToolResponse, error) {
	http.DefaultClient := http.DefaultClient

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonBytes, jsonErr := json.Marshal(requestBody)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", jsonErr))
}

	req, reqErr := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(jsonBytes)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		return err(fmt.Sprintf("API request failed: %v", respErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
}

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	var result map[string]interface{}
	if parseErr := json.Unmarshal(bodyBytes, &result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	// Extract the assistant message content
	if choices, found := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, found := choices[0].(map[string]interface{}); found {
			if message, found := choice["message"].(map[string]interface{}); found {
				if content, found := message["content"].(string); found {
					return ok(content)
				}
			}
		}
	}

	return err("failed to extract response from OpenAI format")

func callAnthropic(apiKey, model, prompt string) (ToolResponse, error) {
	http.DefaultClient := http.DefaultClient

	requestBody := map[string]interface{}{
		"model": model,
		"max_tokens": 1024,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonBytes, jsonErr := json.Marshal(requestBody)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", jsonErr))
}

	req, reqErr := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", strings.NewReader(string(jsonBytes)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		return err(fmt.Sprintf("API request failed: %v", respErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
}

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	var result map[string]interface{}
	if parseErr := json.Unmarshal(bodyBytes, &result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	// Extract the assistant message content
	if content, found := result["content"].([]interface{}); ok && len(content) > 0 {
		if textBlock, found := content[0].(map[string]interface{}); found {
			if text, found := textBlock["text"].(string); found {
				return ok(text)
			}
		}
	}

	return err("failed to extract response from Anthropic format")
}

// HandleListMCPTools lists available tools from an MCP server
func HandleListMCPTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	if serverURL == "" {
		return err("server_url is required")
}

	http.DefaultClient := http.DefaultClient

	// Request tools list from MCP server
	req, reqErr := http.NewRequest("GET", serverURL+"/tools", nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		return err(fmt.Sprintf("API request failed: %v", respErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
}

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	var result struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if parseErr := json.Unmarshal(bodyBytes, &result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse tools list: %v", parseErr))
}

	// Extract tool names
	names := make([]string, len(result.Tools))
	for i, tool := range result.Tools {
		names[i] = tool.Name
	}

	return ok(strings.Join(names, "\n"))
}

// HandleRunAgent runs an agent with the given configuration
func HandleRunAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agentID, _ :=getString(args, "agent_id")
	if agentID == "" {
		return err("agent_id is required")
}

	task, _ :=getString(args, "task")
	if task == "" {
		return err("task is required")
}

	// In a real implementation, this would start an agent process
	// For now, we simulate by returning a success message
	return ok(fmt.Sprintf("Agent %s started with task: %s", agentID, task))
}