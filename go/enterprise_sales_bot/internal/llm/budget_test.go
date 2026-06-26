package llm

import (
	"context"
	"testing"
	"time"
)

// MockProvider implements LLMProvider for testing BudgetAwareProvider.

type MockProvider struct{}

func (m *MockProvider) Generate(ctx context.Context, prompt Prompt) (string, error) {
	// Return a dummy response of fixed length.
	return "dummy response", nil
}

func TestTokenBudget_BasicUsage(t *testing.T) {
	budget := NewTokenBudget(1000, time.Hour, 0.8, nil)
	if !budget.IsWithinBudget() {
		t.Fatalf("budget should be initially within limits")
	}

	// Record 200 tokens usage.
	ok := budget.RecordUsage(200, 0, 0, "model", "prompt", "response", "test")
	if !ok {
		t.Fatalf("recorded usage should be within budget")
	}

	used, total, remaining, percent := budget.GetUsage()
	if used != 200 || total != 1000 || remaining != 800 {
		t.Fatalf("unexpected usage values: used=%d total=%d remaining=%d", used, total, remaining)
	}
	if percent < 0.19 || percent > 0.21 {
		t.Fatalf("percent used out of range: %f", percent)
	}
}

func TestTokenBudget_Exceeded(t *testing.T) {
	budget := NewTokenBudget(500, time.Hour, 0.8, nil)
	budget.RecordUsage(400, 0, 0, "model", "prompt", "response", "test")
	if !budget.IsWithinBudget() {
		t.Fatalf("budget should be within limit after 400/500 used")
	}
	if budget.ShouldWarn() != true {
		t.Fatalf("should warn after exceeding warning threshold (80%%)")
	}

	// Record additional 200 tokens – should exceed budget.
	ok := budget.RecordUsage(200, 0, 0, "model", "prompt", "response", "test")
	if ok {
		t.Fatalf("expected RecordUsage to return false when budget exceeded")
	}
}

func TestBudgetAwareProvider_EnforcesBudget(t *testing.T) {
	// 1k token budget, each mock call estimated ~ (system+user+response)/4 = small, but we force usage.
	budget := NewTokenBudget(100, time.Hour, 0.8, nil)
	tracker := NewDealTokenTracker()
	provider := NewBudgetAwareProvider(&MockProvider{}, budget, tracker, 0.0)

	prompt := Prompt{User: "test prompt", System: "system", MaxTokens: 100}

	// First call should succeed.
	_, err := provider.Generate(context.Background(), prompt)
	if err != nil {
		t.Fatalf("first generate should succeed: %v", err)
	}

	// Exhaust budget manually.
	budget.RecordUsage(100, 0, 0, "model", "prompt", "resp", "test")

	_, err = provider.Generate(context.Background(), prompt)
	if err == nil {
		t.Fatalf("expected error due to budget exceeded")
	}
	if err != ErrBudgetExceeded {
		t.Fatalf("expected ErrBudgetExceeded, got %v", err)
	}
}

func TestDealTokenTracker_RecordsUsage(t *testing.T) {
	tracker := NewDealTokenTracker()
	tracker.RecordUsage(42, 123, "model", "prompt", "response", "test")
	usage := tracker.GetDealUsage(42)
	if usage == nil || usage.DealID != 42 || usage.Tokens != 123 {
		t.Fatalf("tracker returned incorrect usage: %+v", usage)
	}
}
