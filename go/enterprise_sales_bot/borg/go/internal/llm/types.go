package llm

import (
	"context"
	"time"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolSchema struct {
	Type     string      `json:"type"`
	Function FunctionDef `json:"function"`
}

type FunctionDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ChatRequest struct {
	Model       string       `json:"model"`
	Messages    []Message    `json:"messages"`
	Tools       []ToolSchema `json:"tools,omitempty"`
	Temperature float64      `json:"temperature,omitempty"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	TopP        float64      `json:"top_p,omitempty"`
	Stream      bool         `json:"stream,omitempty"`
	MaxRetries  int          `json:"-"`
	Attempt     int          `json:"-"`
	Exclude     []string     `json:"-"`
}

type ChatResponse struct {
	ID        string    `json:"id"`
	Model     string    `json:"model"`
	Choices   []Choice  `json:"choices"`
	Usage     Usage     `json:"usage"`
	CreatedAt time.Time `json:"created_at"`
	ProviderID string   `json:"-"`
	Attempts   int      `json:"-"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ProviderError struct {
	ProviderID string
	StatusCode int
	Body       string
	Retryable  bool
}

func (e *ProviderError) Error() string {
	return "llm: provider " + e.ProviderID + " returned HTTP " + itoa(e.StatusCode) + ": " + e.Body
}

type LLMClient interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
}

type StreamChunk struct {
	Delta *Message
	Done  bool
	Err   error
	Usage *Usage
}

type TierConfig struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	BaseURL      string            `json:"baseUrl"`
	APIKey       string            `json:"apiKey"`
	DefaultModel string            `json:"defaultModel"`
	Timeout      time.Duration     `json:"timeout"`
	Headers      map[string]string `json:"headers"`
	Priority     int               `json:"priority"`
}

func DefaultTierConfigs() []TierConfig {
	return []TierConfig{
		{
			ID:           "nvidia-nim",
			Name:         "NVIDIA NIM",
			BaseURL:      "https://integrate.api.nvidia.com/v1",
			DefaultModel: "meta/llama-3.1-70b-instruct",
			Timeout:      30 * time.Second,
			Priority:     0,
		},
		{
			ID:           "openrouter",
			Name:         "OpenRouter",
			BaseURL:      "https://openrouter.ai/api/v1",
			DefaultModel: "anthropic/claude-sonnet-4",
			Timeout:      45 * time.Second,
			Priority:     1,
		},
		{
			ID:           "lmstudio",
			Name:         "LM Studio (Local)",
			BaseURL:      "http://localhost:1234/v1",
			DefaultModel: "local-model",
			Timeout:      120 * time.Second,
			Priority:     2,
		},
	}
}

func itoa(n int) string {
	if n < 0 {
		return "-" + itoa(-n)
	}
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}
