package communication

import (
	"context"
	"testing"
)

func TestSentimentAnalyzer_Analyze_Positive(t *testing.T) {
	sa := NewSentimentAnalyzer(nil)

	result, err := sa.Analyze(context.Background(), "This looks really interesting! Tell me more about your solution. Let's schedule a demo.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Sentiment != SentimentPositive {
		t.Errorf("expected positive sentiment, got %s", result.Sentiment)
	}
	if result.Score < 30 {
		t.Errorf("expected high score for positive message, got %d", result.Score)
	}
	if result.Confidence < 0.5 {
		t.Errorf("expected confidence >= 0.5, got %f", result.Confidence)
	}
}

func TestSentimentAnalyzer_Analyze_Negative(t *testing.T) {
	sa := NewSentimentAnalyzer(nil)

	result, err := sa.Analyze(context.Background(), "Not interested. This is a waste of my time. Please unsubscribe me.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Sentiment != SentimentNegative {
		t.Errorf("expected negative sentiment, got %s", result.Sentiment)
	}
	if result.Score > -20 {
		t.Errorf("expected low score for negative message, got %d", result.Score)
	}
}

func TestSentimentAnalyzer_Analyze_Neutral(t *testing.T) {
	sa := NewSentimentAnalyzer(nil)

	result, err := sa.Analyze(context.Background(), "Can you send more information about your product? I'll review it with my team.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Sentiment != SentimentNeutral && result.Sentiment != SentimentPositive {
		t.Errorf("expected neutral or positive sentiment, got %s", result.Sentiment)
	}
	if len(result.Keywords) == 0 {
		t.Error("expected keywords to be detected")
	}
}

func TestSentimentAnalyzer_Analyze_Urgency(t *testing.T) {
	sa := NewSentimentAnalyzer(nil)

	result, err := sa.Analyze(context.Background(), "This is urgent. We need a solution immediately. Critical deadline approaching.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Urgency < 0.3 {
		t.Errorf("expected high urgency, got %f", result.Urgency)
	}
}

func TestSentimentAnalyzer_Analyze_Empty(t *testing.T) {
	sa := NewSentimentAnalyzer(nil)

	result, err := sa.Analyze(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Sentiment != SentimentNeutral {
		t.Errorf("expected neutral for empty message, got %s", result.Sentiment)
	}
}

func TestAggregateDealSentiment(t *testing.T) {
	results := []SentimentResult{
		{Sentiment: SentimentPositive, Score: 60, Urgency: 0.5},
		{Sentiment: SentimentPositive, Score: 80, Urgency: 0.3},
		{Sentiment: SentimentNegative, Score: -40, Urgency: 0.8},
		{Sentiment: SentimentNeutral, Score: 0, Urgency: 0.0},
	}

	summary := AggregateDealSentiment(results)
	if summary.TotalMessages != 4 {
		t.Errorf("expected 4 messages, got %d", summary.TotalMessages)
	}
	if summary.MostCommon != "positive" {
		t.Errorf("expected most common to be positive, got %s", summary.MostCommon)
	}
	if summary.Trend == "" {
		t.Error("expected non-empty trend")
	}
}

func TestAggregateDealSentiment_Empty(t *testing.T) {
	summary := AggregateDealSentiment(nil)
	if summary.TotalMessages != 0 {
		t.Errorf("expected 0 for empty, got %d", summary.TotalMessages)
	}
}

func TestSentiment_String(t *testing.T) {
	tests := []struct {
		s    Sentiment
		want string
	}{
		{SentimentPositive, "positive"},
		{SentimentNegative, "negative"},
		{SentimentNeutral, "neutral"},
		{SentimentMixed, "mixed"},
		{SentimentUnknown, "unknown"},
		{Sentiment(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.s.String(); got != tt.want {
			t.Errorf("Sentiment(%d).String() = %q, want %q", tt.s, got, tt.want)
		}
	}
}
