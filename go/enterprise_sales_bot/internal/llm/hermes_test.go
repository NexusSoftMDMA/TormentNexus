package llm

import (
	"context"
	"log"
	"os"
	"testing"
)

func TestHermesLLMProvider_Construct(t *testing.T) {
	provider := NewHermesLLMProvider(HermesConfig{
		BaseURL: "http://localhost:8642",
		APIKey:  "test-key",
		Model:   "free-llm",
	})

	if provider.BaseURL != "http://localhost:8642" {
		t.Errorf("expected BaseURL http://localhost:8642, got %s", provider.BaseURL)
	}
	if provider.Model != "free-llm" {
		t.Errorf("expected Model free-llm, got %s", provider.Model)
	}
	if provider.HTTPClient == nil {
		t.Error("expected HTTPClient to be initialized")
	}
}

func TestHermesLLMProvider_TrailingSlashTrimmed(t *testing.T) {
	provider := NewHermesLLMProvider(HermesConfig{
		BaseURL: "http://localhost:8642/",
		APIKey:  "test-key",
		Model:   "free-llm",
	})

	if provider.BaseURL != "http://localhost:8642" {
		t.Errorf("expected trailing slash trimmed, got %s", provider.BaseURL)
	}
}

func TestHermesLLMProvider_Integration(t *testing.T) {
	url := os.Getenv("HERMES_API_URL")
	key := os.Getenv("HERMES_API_KEY")
	model := os.Getenv("HERMES_MODEL")

	if url == "" || key == "" {
		t.Skip("HERMES_API_URL or HERMES_API_KEY not set, skipping integration test")
	}

	if model == "" {
		model = "free-llm"
	}

	provider := NewHermesLLMProvider(HermesConfig{
		BaseURL: url,
		APIKey:  key,
		Model:   model,
	})

	// Health check
	if err := provider.HealthCheck(context.Background()); err != nil {
		t.Fatalf("Hermes health check failed: %v", err)
	}
	log.Println("Health check passed")

	// Simple generation
	result, err := provider.Generate(context.Background(), Prompt{
		System: "You are a helpful assistant. Respond with exactly one word.",
		User:   "What is the capital of France?",
	})
	if err != nil {
		t.Fatalf("Hermes generation failed: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty response from Hermes")
	}

	log.Printf("Hermes response: %q", result)
}
