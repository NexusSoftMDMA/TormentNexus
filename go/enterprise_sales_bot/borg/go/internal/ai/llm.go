package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMResponse struct {
	Content  string `json:"content"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Usage    struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type Provider interface {
	GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error)
}

type QuotaTracker interface {
    UpdateUsage(provider string, tokens int64, cost float64)
}

var GlobalQuotaTracker QuotaTracker

type OpenAIProvider struct {
	APIKey  string
	BaseURL string
}

func (p *OpenAIProvider) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	if p.BaseURL == "" {
		p.BaseURL = "https://api.openai.com/v1/chat/completions"
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model":    model,
		"messages": messages,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	var payload struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if len(payload.Choices) == 0 {
		return nil, fmt.Errorf("no content returned from OpenAI")
	}

	res := &LLMResponse{
		Content:  payload.Choices[0].Message.Content,
		Provider: "openai",
		Model:    model,
	}
	res.Usage.InputTokens = payload.Usage.PromptTokens
	res.Usage.OutputTokens = payload.Usage.CompletionTokens

    if GlobalQuotaTracker != nil {
        GlobalQuotaTracker.UpdateUsage("openai", int64(res.Usage.InputTokens + res.Usage.OutputTokens), 0.0)
    }

    return res, nil
}

type AnthropicProvider struct {
	APIKey  string
	BaseURL string
}

func (p *AnthropicProvider) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	if p.BaseURL == "" {
		p.BaseURL = "https://api.anthropic.com/v1/messages"
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model":      model,
		"max_tokens": 4096,
		"messages":   messages,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Anthropic API error: %s - %s", resp.Status, string(body))
	}

	var payload struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if len(payload.Content) == 0 {
		return nil, fmt.Errorf("no content returned from Anthropic")
	}

	res := &LLMResponse{
		Content:  payload.Content[0].Text,
		Provider: "anthropic",
		Model:    model,
	}
	res.Usage.InputTokens = payload.Usage.InputTokens
	res.Usage.OutputTokens = payload.Usage.OutputTokens

    if GlobalQuotaTracker != nil {
        GlobalQuotaTracker.UpdateUsage("anthropic", int64(res.Usage.InputTokens + res.Usage.OutputTokens), 0.0)
    }

    return res, nil
}

type GeminiProvider struct {
	APIKey  string
	BaseURL string
}

func (p *GeminiProvider) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	if p.BaseURL == "" {
		p.BaseURL = "https://generativelanguage.googleapis.com/v1beta"
	}

	contents := make([]map[string]interface{}, 0, len(messages))
	var systemInstruction string
	for _, msg := range messages {
		if msg.Role == "system" {
			systemInstruction = msg.Content
			continue
		}
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, map[string]interface{}{
			"role":  role,
			"parts": []map[string]string{{"text": msg.Content}},
		})
	}

	body := map[string]interface{}{
		"contents": contents,
	}
	if systemInstruction != "" {
		body["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]string{{"text": systemInstruction}},
		}
	}

	reqBody, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.BaseURL, model, p.APIKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini API error: %s - %s", resp.Status, string(respBody))
	}

	var payload struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if len(payload.Candidates) == 0 || len(payload.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content returned from Gemini")
	}

	var sb strings.Builder
	for _, part := range payload.Candidates[0].Content.Parts {
		sb.WriteString(part.Text)
	}

	res := &LLMResponse{
		Content:  sb.String(),
		Provider: "google",
		Model:    model,
	}
	res.Usage.InputTokens = payload.UsageMetadata.PromptTokenCount
	res.Usage.OutputTokens = payload.UsageMetadata.CandidatesTokenCount

    if GlobalQuotaTracker != nil {
        GlobalQuotaTracker.UpdateUsage("google", int64(res.Usage.InputTokens + res.Usage.OutputTokens), 0.0)
    }

    return res, nil
}

type DeepSeekProvider struct {
	APIKey string
}

func (p *DeepSeekProvider) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	oai := &OpenAIProvider{
		APIKey:  p.APIKey,
		BaseURL: "https://api.deepseek.com/v1/chat/completions",
	}
	resp, err := oai.GenerateText(ctx, model, messages)
	if err != nil {
		return nil, err
	}
	resp.Provider = "deepseek"
	return resp, nil
}

type OpenRouterProvider struct {
	APIKey string
}

func (p *OpenRouterProvider) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	oai := &OpenAIProvider{
		APIKey:  p.APIKey,
		BaseURL: "https://openrouter.ai/api/v1/chat/completions",
	}
	resp, err := oai.GenerateText(ctx, model, messages)
	if err != nil {
		return nil, err
	}
	resp.Provider = "openrouter"
	return resp, nil
}

type LMStudioProvider struct {
	BaseURL string
}

func (p *LMStudioProvider) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	if p.BaseURL == "" {
		p.BaseURL = "http://localhost:1234/v1/chat/completions"
	}
	oai := &OpenAIProvider{
		APIKey:  "lm-studio",
		BaseURL: p.BaseURL,
	}
	resp, err := oai.GenerateText(ctx, model, messages)
	if err != nil {
		return nil, err
	}
	resp.Provider = "lmstudio"
	return resp, nil
}

type OllamaProvider struct {
	BaseURL string
}

func (p *OllamaProvider) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	if p.BaseURL == "" {
		p.BaseURL = "http://localhost:11434/api/chat"
	}

	body := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   false,
	}
	reqBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		PromptEvalCount int `json:"prompt_eval_count"`
		EvalCount       int `json:"eval_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	res := &LLMResponse{
		Content:  payload.Message.Content,
		Provider: "ollama",
		Model:    model,
	}
	res.Usage.InputTokens = payload.PromptEvalCount
	res.Usage.OutputTokens = payload.EvalCount

	return res, nil
}

var ProviderPriority = []struct {
	EnvVar       string
	ProviderName string
	DefaultModel string
	Factory      func(apiKey string) Provider
}{
	{"ANTHROPIC_API_KEY", "anthropic", "claude-sonnet-4-20250514", func(k string) Provider { return &AnthropicProvider{APIKey: k} }},
	{"GOOGLE_AI_STUDIO_API_KEY", "google-ai-studio", "gemini-2.5-flash", func(k string) Provider { return &GeminiProvider{APIKey: k} }},
	{"GOOGLE_API_KEY", "google", "gemini-2.5-flash", func(k string) Provider { return &GeminiProvider{APIKey: k} }},
	{"GEMINI_API_KEY", "google", "gemini-2.5-flash", func(k string) Provider { return &GeminiProvider{APIKey: k} }},
	{"OPENAI_API_KEY", "openai", "gpt-4o", func(k string) Provider { return &OpenAIProvider{APIKey: k} }},
	{"DEEPSEEK_API_KEY", "deepseek", "deepseek-chat", func(k string) Provider { return &DeepSeekProvider{APIKey: k} }},
	{"OPENROUTER_API_KEY", "openrouter", "openrouter/free", func(k string) Provider { return &OpenRouterProvider{APIKey: k} }},
	{"", "lmstudio", "local-model", func(k string) Provider { return &LMStudioProvider{} }},
	{"", "ollama", "gemma:2b", func(k string) Provider { return &OllamaProvider{} }},
}

type providerSelection struct {
	EnvVar       string
	ProviderName string
	DefaultModel string
	Factory      func(apiKey string) Provider
	APIKey       string
}

func resolveProviderSelection() (providerSelection, bool) {
	for _, entry := range ProviderPriority {
		if entry.EnvVar == "" { continue }
		if key := os.Getenv(entry.EnvVar); key != "" {
			return providerSelection{
				EnvVar:       entry.EnvVar,
				ProviderName: entry.ProviderName,
				DefaultModel: entry.DefaultModel,
				Factory:      entry.Factory,
				APIKey:       key,
			}, true
		}
	}
	return providerSelection{}, false
}

func AutoRoute(ctx context.Context, messages []Message) (*LLMResponse, error) {
	selection, ok := resolveProviderSelection()
	if !ok {
		return nil, fmt.Errorf("no LLM provider configured")
	}
	return selection.Factory(selection.APIKey).GenerateText(ctx, selection.DefaultModel, messages)
}

func AutoRouteWithModel(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	selection, ok := resolveProviderSelection()
	if !ok {
		return nil, fmt.Errorf("no LLM provider configured")
	}
	if model == "" {
		model = selection.DefaultModel
	}
	return selection.Factory(selection.APIKey).GenerateText(ctx, model, messages)
}

func ListConfiguredProviders() []string {
	var configured []string
	seen := map[string]struct{}{}
	for _, entry := range ProviderPriority {
		if entry.EnvVar != "" && os.Getenv(entry.EnvVar) == "" {
			continue
		}
		if _, ok := seen[entry.ProviderName]; ok {
			continue
		}
		seen[entry.ProviderName] = struct{}{}
		configured = append(configured, entry.ProviderName)
	}
	return configured
}
