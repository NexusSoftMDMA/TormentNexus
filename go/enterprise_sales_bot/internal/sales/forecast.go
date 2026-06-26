package sales

import (
	"math"
	"sort"
	"sync"

	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
)

// DealForecast represents the predicted outcome for a deal.
type DealForecast struct {
	DealID          int64   `json:"deal_id"`
	WinProbability  float64 `json:"win_probability"`  // 0.0 – 1.0
	ExpectedValue   float64 `json:"expected_value"`    // probability * quoted pricing
	PredictedStage  string  `json:"predicted_stage"`   // next likely stage
	TimeToCloseDays int     `json:"time_to_close_days"`
	Confidence      string  `json:"confidence"`        // high, medium, low
	RiskFactors     []RiskFactor `json:"risk_factors"`
	Recommendation  string  `json:"recommendation"`
}

// RiskFactor describes something that could impact the deal.
type RiskFactor struct {
	Factor   string `json:"factor"`
	Severity string `json:"severity"` // high, medium, low
	Detail   string `json:"detail"`
}

// ForecastingEngine predicts deal outcomes using historical patterns.
type ForecastingEngine struct {
	mu          sync.RWMutex
	historicals []HistoricalDeal // learned from closed deals
	winRateBySource map[string]float64
	avgTimeByStage  map[string]float64 // avg days per stage
}

// HistoricalDeal stores data from a closed deal for pattern learning.
type HistoricalDeal struct {
	DealID          int64
	Source          string // e.g., "hn", "linkedin", "github"
	TotalDays       int
	InteractionCount int
	AvgSentiment    float64
	StageCount      int // number of stages traversed
	Won             bool
}

// NewForecastingEngine creates a forecasting engine.
func NewForecastingEngine() *ForecastingEngine {
	return &ForecastingEngine{
		historicals:     make([]HistoricalDeal, 0),
		winRateBySource: make(map[string]float64),
		avgTimeByStage:  make(map[string]float64),
	}
}

// LearnFromDeal records a closed deal for future pattern matching.
func (fe *ForecastingEngine) LearnFromDeal(deal HistoricalDeal) {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	fe.historicals = append(fe.historicals, deal)
	fe.recalculate()
}

// recalculate updates aggregate statistics from historical data.
func (fe *ForecastingEngine) recalculate() {
	// Win rate by source
	winCounts := make(map[string]int)
	totalCounts := make(map[string]int)
	for _, h := range fe.historicals {
		totalCounts[h.Source]++
		if h.Won {
			winCounts[h.Source]++
		}
	}
	for source, total := range totalCounts {
		if total > 0 {
			fe.winRateBySource[source] = float64(winCounts[source]) / float64(total)
		}
	}

	// Average time per stage (placeholder — in production this would be per-stage tracking)
	fe.avgTimeByStage["discovered"] = 3.0
	fe.avgTimeByStage["researched"] = 5.0
	fe.avgTimeByStage["outreach_sent"] = 7.0
	fe.avgTimeByStage["engaged"] = 14.0
	fe.avgTimeByStage["negotiating"] = 21.0
}

// Forecast produces a prediction for a deal based on current state and history.
func (fe *ForecastingEngine) Forecast(dealID int64, currentStage string, daysInStage int,
	interactionCount int, sentimentResults []communication.SentimentResult,
	quotedPricing float64, source string) DealForecast {

	fe.mu.RLock()
	defer fe.mu.RUnlock()

	forecast := DealForecast{
		DealID:     dealID,
		RiskFactors: make([]RiskFactor, 0),
	}

	// 1. Base probability from source win rate
	sourceWinRate := fe.winRateBySource[source]
	if sourceWinRate == 0 {
		sourceWinRate = 0.25 // default: 25% base
	}

	// 2. Stage-based probability
	stageProb := stageWinProbability(currentStage)

	// 3. Time-based penalty: deals stuck too long are less likely to close
	expectedDays := expectedDaysInStage(currentStage)
	timeRatio := 1.0
	if expectedDays > 0 {
		timeRatio = float64(daysInStage) / expectedDays
		if timeRatio < 0.5 {
			timeRatio = 0.5 // bonus for moving fast
		} else if timeRatio > 3.0 {
			timeRatio = 3.0 // penalty for being stuck
		}
	}
	timePenalty := 1.0 - (timeRatio-0.5)/2.5 // normalized: 0.8 – 1.0 for fast, 0.0 – 0.8 for stuck
	if timePenalty < 0 {
		timePenalty = 0
	}

	// 4. Interaction quality (using sentiment)
	interactionScore := 0.5 // neutral
	if len(sentimentResults) > 0 {
		var avgScore float64
		for _, r := range sentimentResults {
			avgScore += float64(r.Score)
		}
		avgScore /= float64(len(sentimentResults))
		interactionScore = (avgScore + 100) / 200.0 // normalize -100..100 to 0..1
	}

	// 5. Interaction quantity bonus
	interactionBonus := 1.0
	if interactionCount == 0 {
		interactionBonus = 0.7 // no interactions = low engagement
	} else if interactionCount >= 5 {
		interactionBonus = 1.1 // good engagement
	} else if interactionCount >= 3 {
		interactionBonus = 1.0
	}

	// Combine factors
	rawProb := sourceWinRate * 0.3 + stageProb * 0.35 + timePenalty * 0.15 + interactionScore * 0.2
	rawProb *= interactionBonus

	// Clamp to 0–1
	if rawProb < 0 {
		rawProb = 0
	} else if rawProb > 1 {
		rawProb = 1
	}

	forecast.WinProbability = math.Round(rawProb*100) / 100
	forecast.ExpectedValue = forecast.WinProbability * quotedPricing

	// Predicted next stage
	forecast.PredictedStage = nextStage(currentStage)

	// Time to close estimate
	forecast.TimeToCloseDays = estimateTimeToClose(currentStage, daysInStage)

	// Confidence
	switch {
	case interactionCount >= 3 && len(sentimentResults) >= 2:
		forecast.Confidence = "high"
	case interactionCount >= 2 && len(sentimentResults) >= 1:
		forecast.Confidence = "medium"
	default:
		forecast.Confidence = "low"
	}

	// Risk factors
	if timeRatio > 2.0 {
		forecast.RiskFactors = append(forecast.RiskFactors, RiskFactor{
			Factor:   "stalled_progress",
			Severity: "high",
			Detail:   "Deal has been in stage longer than expected (%.0f vs %.0f days)",
		})
	}
	if interactionCount == 0 {
		forecast.RiskFactors = append(forecast.RiskFactors, RiskFactor{
			Factor:   "no_engagement",
			Severity: "high",
			Detail:   "No interactions recorded for this deal",
		})
	}
	if interactionScore < 0.3 && len(sentimentResults) > 0 {
		forecast.RiskFactors = append(forecast.RiskFactors, RiskFactor{
			Factor:   "negative_sentiment",
			Severity: "medium",
			Detail:   "Recent interactions show negative sentiment trend",
		})
	}
	if sourceWinRate < 0.2 && len(fe.historicals) > 0 {
		forecast.RiskFactors = append(forecast.RiskFactors, RiskFactor{
			Factor:   "low_source_win_rate",
			Severity: "medium",
			Detail:   "Deals from this source have historically low win rates",
		})
	}

	// Recommendation
	switch {
	case forecast.WinProbability >= 0.7:
		forecast.Recommendation = "High probability — prioritize for close, offer limited-time incentive"
	case forecast.WinProbability >= 0.4:
		forecast.Recommendation = "Moderate probability — increase engagement cadence, address any objections"
	default:
		forecast.Recommendation = "Low probability — consider re-qualification or move to nurture sequence"
	}

	return forecast
}

// stageWinProbability returns baseline win probability for a stage.
func stageWinProbability(stage string) float64 {
	switch stage {
	case "discovered":
		return 0.10
	case "researched":
		return 0.15
	case "outreach_sent":
		return 0.20
	case "engaged":
		return 0.40
	case "negotiating":
		return 0.65
	case "closed_won":
		return 1.0
	case "closed_lost":
		return 0.0
	default:
		return 0.15
	}
}

// expectedDaysInStage returns the expected number of days a deal should spend in a stage.
func expectedDaysInStage(stage string) float64 {
	switch stage {
	case "discovered":
		return 2.0
	case "researched":
		return 3.0
	case "outreach_sent":
		return 5.0
	case "engaged":
		return 10.0
	case "negotiating":
		return 14.0
	default:
		return 7.0
	}
}

// nextStage returns the next logical stage for progression.
func nextStage(current string) string {
	switch current {
	case "discovered":
		return "researched"
	case "researched":
		return "outreach_sent"
	case "outreach_sent":
		return "engaged"
	case "engaged":
		return "negotiating"
	case "negotiating":
		return "closed_won"
	default:
		return "engaged"
	}
}

// estimateTimeToClose predicts remaining days to close.
func estimateTimeToClose(stage string, daysInStage int) int {
	totalExpected := expectedDaysInStage("discovered") +
		expectedDaysInStage("researched") +
		expectedDaysInStage("outreach_sent") +
		expectedDaysInStage("engaged") +
		expectedDaysInStage("negotiating")

	elapsed := stageElapsedDays(stage, daysInStage)
	remaining := int(totalExpected - elapsed)
	if remaining < 1 {
		remaining = 1
	}
	return remaining
}

// stageElapsedDays calculates total days elapsed from start to current stage.
func stageElapsedDays(stage string, daysInStage int) float64 {
	base := 0.0
	switch stage {
	case "discovered":
		base = 0
	case "researched":
		base = expectedDaysInStage("discovered")
	case "outreach_sent":
		base = expectedDaysInStage("discovered") + expectedDaysInStage("researched")
	case "engaged":
		base = expectedDaysInStage("discovered") + expectedDaysInStage("researched") + expectedDaysInStage("outreach_sent")
	case "negotiating":
		base = expectedDaysInStage("discovered") + expectedDaysInStage("researched") + expectedDaysInStage("outreach_sent") + expectedDaysInStage("engaged")
	case "closed_won", "closed_lost":
		base = expectedDaysInStage("discovered") + expectedDaysInStage("researched") + expectedDaysInStage("outreach_sent") + expectedDaysInStage("engaged") + expectedDaysInStage("negotiating")
	}
	return base + float64(daysInStage)
}

// PipelineSummary provides an aggregate view of the sales pipeline.
type PipelineSummary struct {
	TotalDeals        int     `json:"total_deals"`
	TotalValue        float64 `json:"total_value"`
	ExpectedRevenue   float64 `json:"expected_revenue"`
	AvgWinProbability float64 `json:"avg_win_probability"`
	ByStage           map[string]StageSummary `json:"by_stage"`
	AtRiskDeals       int     `json:"at_risk_deals"`
}

// StageSummary aggregates deals in a single stage.
type StageSummary struct {
	Count     int     `json:"count"`
	TotalValue float64 `json:"total_value"`
}

// SummarizePipeline aggregates forecast data across all deals.
func SummarizePipeline(forecasts []DealForecast) PipelineSummary {
	summary := PipelineSummary{
		ByStage: make(map[string]StageSummary),
	}

	var totalProb float64
	for _, f := range forecasts {
		summary.TotalDeals++
		summary.TotalValue += f.ExpectedValue / (f.WinProbability + 0.01) // approximate total value
		summary.ExpectedRevenue += f.ExpectedValue
		totalProb += f.WinProbability

		// By stage
		stage := f.PredictedStage
		s := summary.ByStage[stage]
		s.Count++
		s.TotalValue += f.ExpectedValue
		summary.ByStage[stage] = s

		// At-risk detection
		if f.WinProbability < 0.2 {
			summary.AtRiskDeals++
		}
	}

	if summary.TotalDeals > 0 {
		summary.AvgWinProbability = totalProb / float64(summary.TotalDeals)
	}

	return summary
}

// PercentileForecast calculates the P10, P50, P90 revenue forecast.
func (fe *ForecastingEngine) PercentileForecast(forecasts []DealForecast) (p10, p50, p90 float64) {
	if len(forecasts) == 0 {
		return 0, 0, 0
	}

	sorted := make([]float64, len(forecasts))
	for i, f := range forecasts {
		sorted[i] = f.ExpectedValue
	}
	sort.Float64s(sorted)

	n := len(sorted)
	p10 = sorted[int(float64(n)*0.1)]
	if p10 < 0 {
		p10 = 0
	}
	p50 = sorted[int(float64(n)*0.5)]
	p90 = sorted[int(float64(n)*0.9)]
	return
}
