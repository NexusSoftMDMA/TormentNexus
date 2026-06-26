package communication

import (
	"context"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestNewObjectionLibrary(t *testing.T) {
	lib := NewObjectionLibrary()
	if lib == nil {
		t.Fatal("expected non-nil library")
	}

	stats := lib.Statistics()
	count, ok := stats["objection_count"].(int)
	if !ok || count == 0 {
		t.Errorf("expected non-zero objection count, got %d", count)
	}
	respCount, ok := stats["response_count"].(int)
	if !ok || respCount == 0 {
		t.Errorf("expected non-zero response count, got %d", respCount)
	}
}

func TestMatchObjection_PricingMatch(t *testing.T) {
	lib := NewObjectionLibrary()
	ctx := context.Background()

	text := "This is too expensive for us. We don't have the budget right now."
	sentiment := SentimentResult{
		Sentiment:  SentimentNegative,
		Confidence: 0.8,
		Score:      -40,
	}

	result := lib.MatchObjection(ctx, text, sentiment, db.StateNegotiating)
	if result == nil {
		t.Fatal("expected a match, got nil")
	}
	if result.Objection.Category != CategoryPricing {
		t.Errorf("expected pricing category, got %s", result.Objection.Category)
	}
	if result.Score <= 0 {
		t.Errorf("expected positive score, got %f", result.Score)
	}
	if result.Response.ObjectionID != result.Objection.ID {
		t.Errorf("response objection_id %s does not match objection id %s",
			result.Response.ObjectionID, result.Objection.ID)
	}
}

func TestMatchObjection_SecurityMatch(t *testing.T) {
	lib := NewObjectionLibrary()
	ctx := context.Background()

	text := "We have strict data privacy requirements and need to review your security posture."
	sentiment := SentimentResult{
		Sentiment:  SentimentNeutral,
		Confidence: 0.7,
	}

	result := lib.MatchObjection(ctx, text, sentiment, db.StateEngaged)
	if result == nil {
		t.Fatal("expected a match, got nil")
	}
	if result.Objection.Category != CategorySecurity {
		t.Errorf("expected security category, got %s", result.Objection.Category)
	}
}

func TestMatchObjection_NoMatch(t *testing.T) {
	lib := NewObjectionLibrary()
	ctx := context.Background()

	text := "Sounds great! When can we start the pilot?"
	sentiment := SentimentResult{
		Sentiment:  SentimentPositive,
		Confidence: 0.9,
	}

	result := lib.MatchObjection(ctx, text, sentiment, db.StateOutreachSent)
	if result != nil {
		t.Errorf("expected no match for positive message, got category %s", result.Objection.Category)
	}
}

func TestMatchObjection_VendorLockIn(t *testing.T) {
	lib := NewObjectionLibrary()
	ctx := context.Background()

	text := "We're worried about getting locked into a proprietary platform. What's our exit strategy?"
	sentiment := SentimentResult{
		Sentiment:  SentimentNegative,
		Confidence: 0.75,
	}

	result := lib.MatchObjection(ctx, text, sentiment, db.StateEngaged)
	if result == nil {
		t.Fatal("expected a match for vendor lock-in")
	}
	if result.Objection.Category != CategoryVendorLockIn {
		t.Errorf("expected vendor_lock_in category, got %s", result.Objection.Category)
	}
}

func TestMatchObjection_Competition(t *testing.T) {
	lib := NewObjectionLibrary()
	ctx := context.Background()

	text := "We're already evaluating LangChain and are happy with their solution."
	sentiment := SentimentResult{
		Sentiment:  SentimentNeutral,
		Confidence: 0.8,
	}

	result := lib.MatchObjection(ctx, text, sentiment, db.StateNegotiating)
	if result == nil {
		t.Fatal("expected a match for competition")
	}
	if result.Objection.Category != CategoryCompetition {
		t.Errorf("expected competition category, got %s", result.Objection.Category)
	}
}

func TestMatchObjection_Authority(t *testing.T) {
	lib := NewObjectionLibrary()
	ctx := context.Background()

	text := "I need to check with my VP of Engineering before making any decision."
	sentiment := SentimentResult{
		Sentiment:  SentimentNeutral,
		Confidence: 0.6,
	}

	result := lib.MatchObjection(ctx, text, sentiment, db.StateEngaged)
	if result == nil {
		t.Fatal("expected a match for authority objection")
	}
	if result.Objection.Category != CategoryAuthority {
		t.Errorf("expected authority category, got %s", result.Objection.Category)
	}
}

func TestRecordOutcome_Success(t *testing.T) {
	lib := NewObjectionLibrary()
	ctx := context.Background()

	// First, find a match to get a response ID
	text := "This is too expensive."
	sentiment := SentimentResult{
		Sentiment:  SentimentNegative,
		Confidence: 0.8,
	}
	result := lib.MatchObjection(ctx, text, sentiment, db.StateNegotiating)
	if result == nil {
		t.Fatal("expected a pricing match")
	}

	initialRate := result.Response.SuccessRate
	initialUsed := result.Response.TimesUsed

	// Record a successful outcome
	lib.RecordOutcome(result.Response.ID, true)

	// Check updated stats
	stats := lib.Statistics()
	totalUsed, _ := stats["total_times_used"].(int)
	if totalUsed != initialUsed+1 {
		t.Errorf("expected %d total uses, got %d", initialUsed+1, totalUsed)
	}
	_ = initialRate
}

func TestRecordOutcome_Failure(t *testing.T) {
	lib := NewObjectionLibrary()

	// Get a response and record failure
	ctx := context.Background()
	text := "We have strict data privacy requirements."
	sentiment := SentimentResult{
		Sentiment:  SentimentNeutral,
		Confidence: 0.7,
	}
	result := lib.MatchObjection(ctx, text, sentiment, db.StateEngaged)
	if result == nil {
		t.Skip("no match available for this test")
	}

	lib.RecordOutcome(result.Response.ID, false)

	// Verify it was tracked (success rate should decrease or stay same)
	for _, resp := range lib.responses {
		if resp.ID == result.Response.ID {
			if resp.TimesUsed != 1 {
				t.Errorf("expected 1 usage, got %d", resp.TimesUsed)
			}
			if resp.SuccessRate > 0.5 {
				t.Errorf("expected low success rate after failure, got %f", resp.SuccessRate)
			}
			break
		}
	}
}

func TestLoadJSON_CustomData(t *testing.T) {
	lib := NewObjectionLibrary()

	customData := `{
		"objections": [
			{
				"id": "obj_custom_test",
				"category": "pricing",
				"title": "Custom Objection",
				"patterns": ["custom pattern"],
				"keywords": ["custom"],
				"urgency": 0.5,
				"priority": 50
			}
		],
		"responses": [
			{
				"id": "resp_custom_test",
				"objection_id": "obj_custom_test",
				"text": "Custom response",
				"approach": "value",
				"use_cases": ["*"],
				"success_rate": 0.5,
				"times_used": 0
			}
		]
	}`

	if err := lib.LoadJSON([]byte(customData)); err != nil {
		t.Fatalf("unexpected error loading custom data: %v", err)
	}

	stats := lib.Statistics()
	count, _ := stats["objection_count"].(int)
	if count != 1 {
		t.Errorf("expected 1 objection, got %d", count)
	}
}

func TestMatchObjection_TimingNotNow(t *testing.T) {
	lib := NewObjectionLibrary()
	ctx := context.Background()

	text := "Now is not a good time. We're too busy with other priorities."
	sentiment := SentimentResult{
		Sentiment:  SentimentNegative,
		Confidence: 0.7,
	}

	result := lib.MatchObjection(ctx, text, sentiment, db.StateOutreachSent)
	if result == nil {
		t.Fatal("expected a match for timing objection")
	}
	if result.Objection.Category != CategoryTiming {
		t.Errorf("expected timing category, got %s", result.Objection.Category)
	}
}

func TestMatchObjection_IntegrationMismatch(t *testing.T) {
	lib := NewObjectionLibrary()
	ctx := context.Background()

	text := "This solution doesn't work with our current infrastructure. We use a custom Kubernetes setup that's not compatible with most off-the-shelf tools."
	sentiment := SentimentResult{
		Sentiment:  SentimentNeutral,
		Confidence: 0.65,
	}

	result := lib.MatchObjection(ctx, text, sentiment, db.StateEngaged)
	if result == nil {
		t.Fatal("expected a match for integration objection")
	}
	if result.Objection.Category != CategoryIntegration {
		t.Errorf("expected integration category, got %s", result.Objection.Category)
	}
}
