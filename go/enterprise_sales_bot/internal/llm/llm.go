package llm

import (
	"context"
	"fmt"
)

// Prompt represents the input to an LLM.
type Prompt struct {
	System   string
	User     string
	MaxTokens int
}

// LLMProvider defines the interface for interacting with various large language models.
type LLMProvider interface {
	Generate(ctx context.Context, prompt Prompt) (string, error)
}

// MockLLMProvider simulates an LLM for testing.
type MockLLMProvider struct{}

func (m *MockLLMProvider) Generate(ctx context.Context, prompt Prompt) (string, error) {
	return fmt.Sprintf("[MOCK LLM RESPONSE based on: %s]", prompt.User), nil
}
