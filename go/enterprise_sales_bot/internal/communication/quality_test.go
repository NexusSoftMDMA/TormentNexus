package communication

import (
	"context"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

// MockLLMProviderForQuality implements llm.LLMProvider for testing.
type MockLLMProviderForQuality struct{}

func (m *MockLLMProviderForQuality) Generate(ctx context.Context, prompt llm.Prompt) (string, error) {
	return `{"overall": 75, "is_relevant": true, "tone": "professional", "issues": [], "suggestions": []}`, nil
}

func TestQualityScorer_Evaluate_Personalized(t *testing.T) {
	qs := NewQualityScorer(nil, 60)

	score, err := qs.Evaluate(context.Background(),
		"TormentNexus for Acme Corp",
		"Hi John, I noticed Acme Corp is building AI infrastructure. Our platform helps with orchestration. Let me know if you're interested!",
		"John",
		"Acme Corp",
		"CTO",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !score.IsPersonalized {
		t.Error("expected message to be marked as personalized")
	}
	if !score.IsRelevant {
		t.Error("expected message to be marked as relevant")
	}
	if !score.HasCTAsk {
		t.Error("expected message to have a CTA")
	}
	if score.Overall < 60 {
		t.Errorf("expected overall score >= 60, got %d", score.Overall)
	}
}

func TestQualityScorer_Evaluate_Generic(t *testing.T) {
	qs := NewQualityScorer(nil, 60)

	score, err := qs.Evaluate(context.Background(),
		"Quick Question",
		"Hi, we have a product. Want to buy?",
		"",
		"",
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if score.IsPersonalized {
		t.Error("expected generic message to not be marked as personalized")
	}
	if score.Overall >= 60 {
		t.Errorf("expected overall score < 60 for generic message, got %d", score.Overall)
	}
	if len(score.Issues) == 0 {
		t.Error("expected issues to be identified")
	}
}

func TestQualityScorer_Evaluate_PushyTone(t *testing.T) {
	qs := NewQualityScorer(nil, 60)

	score, err := qs.Evaluate(context.Background(),
		"Act Now!",
		"You MUST act now! This is urgent! Don't miss this exclusive opportunity!",
		"",
		"",
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if score.Tone != "pushy" {
		t.Errorf("expected tone 'pushy', got %q", score.Tone)
	}
	if score.Overall >= 60 {
		t.Errorf("expected pushy message to score < 60, got %d", score.Overall)
	}
}

func TestQualityScorer_Evaluate_NoCTA(t *testing.T) {
	qs := NewQualityScorer(nil, 60)

	score, err := qs.Evaluate(context.Background(),
		"Hello",
		"This is a message about our product. It has no call to action.",
		"",
		"",
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if score.HasCTAsk {
		t.Error("expected message without CTA to be flagged")
	}
	hasCTAIssue := false
	for _, issue := range score.Issues {
		if issue == "No clear call to action" {
			hasCTAIssue = true
			break
		}
	}
	if !hasCTAIssue {
		t.Error("expected 'No clear call to action' in issues")
	}
}

func TestQualityScorer_Evaluate_ShortMessage(t *testing.T) {
	qs := NewQualityScorer(nil, 60)

	score, err := qs.Evaluate(context.Background(),
		"Hi",
		"Buy our product.",
		"",
		"",
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasShortIssue := false
	for _, issue := range score.Issues {
		if issue == "Message is too short (<30 words)" {
			hasShortIssue = true
			break
		}
	}
	if !hasShortIssue {
		t.Error("expected short message to be flagged")
	}
}

func TestQualityScorer_IsPassing(t *testing.T) {
	qs := NewQualityScorer(nil, 60)

	if !qs.IsPassing(QualityScore{Overall: 75}) {
		t.Error("expected score 75 to pass threshold 60")
	}
	if qs.IsPassing(QualityScore{Overall: 59}) {
		t.Error("expected score 59 to fail threshold 60")
	}
}

func TestQualityScorer_ScoreAndLog(t *testing.T) {
	qs := NewQualityScorer(nil, 60)

	score, passing := qs.ScoreAndLog(context.Background(),
		"TormentNexus for Acme Corp",
		"Hi John, I noticed Acme Corp is building AI infrastructure with orchestration challenges. Our platform helps. Let me know your thoughts!",
		"John",
		"Acme Corp",
		"CTO",
	)

	if !passing {
		t.Error("expected personalized message to pass")
	}
	if score.Overall < 60 {
		t.Errorf("expected score >= 60, got %d", score.Overall)
	}
}

func TestQualityScorer_WithLLMProvider(t *testing.T) {
	mockLLM := &MockLLMProviderForQuality{}
	qs := NewQualityScorer(mockLLM, 60)

	score, err := qs.Evaluate(context.Background(),
		"Test",
		"Test body with infrastructure and orchestration keywords.",
		"",
		"",
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// LLM provider should be called but we don't parse response in current impl
	// Just verify it doesn't crash
	if score.Overall == 0 {
		t.Error("expected non-zero score")
	}
}