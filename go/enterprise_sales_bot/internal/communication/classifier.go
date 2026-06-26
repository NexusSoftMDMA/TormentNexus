package communication

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

// MockIntentClassifier simulates intent classification for inbound text.
type MockIntentClassifier struct{}

func (m *MockIntentClassifier) Classify(ctx context.Context, text string) (Intent, error) {
	slog.Info(fmt.Sprintf("MockIntentClassifier: Classifying text: %s", text))

	// Simple heuristic-based mock classification
	if containsAny(text, "pricing", "cost", "license") {
		return IntentPricing, nil
	}
	if containsAny(text, "architecture", "scaling", "how does it work") {
		return IntentTechnical, nil
	}
	if containsAny(text, "no", "not interested", "stop") {
		return IntentObjection, nil
	}

	return IntentTechnical, nil
}

func containsAny(text string, keywords ...string) bool {
	for _, kw := range keywords {
		if contains(text, kw) {
			return true
		}
	}
	return false
}

func contains(text, kw string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(kw))
}

// LLMIntentClassifier uses an LLM to categorize inbound messages.
type LLMIntentClassifier struct {
	llm llm.LLMProvider
}

// NewLLMIntentClassifier creates a new LLM-based classifier.
func NewLLMIntentClassifier(provider llm.LLMProvider) *LLMIntentClassifier {
	return &LLMIntentClassifier{llm: provider}
}

func (c *LLMIntentClassifier) Classify(ctx context.Context, text string) (Intent, error) {
	prompt := llm.Prompt{
		System:	"You are a sales intent classifier. Classify the user's message into one of: Technical, Pricing, Objection, MeetingRequest, FollowUp, Spam, or Unknown.",
		User:	fmt.Sprintf("Classify this message: %s", text),
	}

	resp, err := c.llm.Generate(ctx, prompt)
	if err != nil {
		return IntentUnknown, err
	}

	cleanResp := strings.TrimSpace(resp)
	// Simple matching logic
	switch {
	case strings.Contains(cleanResp, "Technical"):
		return IntentTechnical, nil
	case strings.Contains(cleanResp, "Pricing"):
		return IntentPricing, nil
	case strings.Contains(cleanResp, "Objection"):
		return IntentObjection, nil
	case strings.Contains(cleanResp, "MeetingRequest"):
		return IntentMeetingRequest, nil
	case strings.Contains(cleanResp, "FollowUp"):
		return IntentFollowUp, nil
	case strings.Contains(cleanResp, "Spam"):
		return IntentSpam, nil
	}

	return IntentUnknown, nil
}
