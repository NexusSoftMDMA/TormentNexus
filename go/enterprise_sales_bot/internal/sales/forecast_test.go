package sales

import (
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
)

func TestForecastingEngine_Basic(t *testing.T) {
	fe := NewForecastingEngine()

	// Add historical deals
	fe.LearnFromDeal(HistoricalDeal{DealID: 1, Source: "hn", TotalDays: 20, InteractionCount: 4, AvgSentiment: 0.7, StageCount: 5, Won: true})
	fe.LearnFromDeal(HistoricalDeal{DealID: 2, Source: "linkedin", TotalDays: 30, InteractionCount: 2, AvgSentiment: -0.4, StageCount: 5, Won: false})
	fe.LearnFromDeal(HistoricalDeal{DealID: 3, Source: "hn", TotalDays: 25, InteractionCount: 3, AvgSentiment: 0.2, StageCount: 5, Won: true})

	// Simulate current deal state with positive sentiment
	sentimentResults := []communication.SentimentResult{{Sentiment: communication.SentimentPositive, Score: 70, Urgency: 0.3}, {Sentiment: communication.SentimentPositive, Score: 80, Urgency: 0.2}}

	forecast := fe.Forecast(999, "engaged", 6, 3, sentimentResults, 50000, "hn")

	if forecast.DealID != 999 {
		t.Fatalf("unexpected deal ID: %d", forecast.DealID)
	}
	if forecast.WinProbability < 0.5 {
		t.Fatalf("expected win probability >= 0.5 for engaged deal with positive sentiment, got %f", forecast.WinProbability)
	}
	if forecast.PredictedStage != "negotiating" {
		t.Fatalf("expected next stage negotiating, got %s", forecast.PredictedStage)
	}
	if forecast.ExpectedValue <= 0 {
		t.Fatalf("expected positive expected value")
	}
	if forecast.Confidence != "high" {
		t.Fatalf("expected high confidence, got %s", forecast.Confidence)
	}
	if len(forecast.RiskFactors) != 0 {
		t.Fatalf("expected no risk factors for healthy deal, got %d", len(forecast.RiskFactors))
	}
	if forecast.Recommendation == "" {
		t.Fatalf("expected recommendation text")
	}
}

func TestForecastingEngine_AtRisk(t *testing.T) {
	fe := NewForecastingEngine()

	// Simulate a deal stuck in negotiating with negative sentiment
	sentimentResults := []communication.SentimentResult{{Sentiment: communication.SentimentNegative, Score: -80, Urgency: 0.9}}

	forecast := fe.Forecast(1001, "negotiating", 30, 1, sentimentResults, 80000, "linkedin")

	if forecast.WinProbability >= 0.4 {
		t.Fatalf("expected low win probability for stuck negotiating deal, got %f", forecast.WinProbability)
	}
	if forecast.Confidence != "low" {
		t.Fatalf("expected low confidence for low engagement, got %s", forecast.Confidence)
	}
	if len(forecast.RiskFactors) == 0 {
		t.Fatalf("expected risk factors for at‑risk deal")
	}
}
