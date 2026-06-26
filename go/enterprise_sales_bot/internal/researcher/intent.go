package researcher

import (
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// IntentSignal describes a single point of interest for scoring lead intent.
type IntentSignal struct {
	Type   string // "hiring", "github", "blog", "competitor"
	Weight int    // Base score multiplier
	Value  string // Raw signal data
}

// IntentScore aggregates signals into a normalized intent score (0-100).
type IntentScore struct {
	Score   int
	Signals []IntentSignal
}

// IntentAggregator unifies disparate signals from a company's profile
// into a singular intent score that prioritizes outreach execution.
type IntentAggregator struct{}

// Aggregate synthesizes a comprehensive intent score based on available data.
func (a *IntentAggregator) Aggregate(company db.Company, dossier string) IntentScore {
	var signals []IntentSignal
	totalScore := 0

	// 1. Hiring Signals
	for _, hiring := range company.HiringSignals {
		lowerHiring := strings.ToLower(hiring)
		weight := 5 // Base for generic hiring

		// Tech-specific roles score higher
		if strings.Contains(lowerHiring, "ai") || strings.Contains(lowerHiring, "ml") || strings.Contains(lowerHiring, "machine learning") {
			weight = 15
		}
		if strings.Contains(lowerHiring, "infrastructure") || strings.Contains(lowerHiring, "platform") || strings.Contains(lowerHiring, "orchestration") {
			weight = 20
		}

		signals = append(signals, IntentSignal{
			Type:   "hiring",
			Weight: weight,
			Value:  hiring,
		})
		totalScore += weight
	}

	// 2. Dossier Analysis (GitHub, Blogs, Competitors)
	lowerDossier := strings.ToLower(dossier)

	// Github/Code Signals
	if strings.Contains(lowerDossier, "github insight") {
		weight := 10
		if strings.Contains(lowerDossier, "orchestration logic") || strings.Contains(lowerDossier, "state management") {
			weight = 25
		}
		signals = append(signals, IntentSignal{
			Type:   "github",
			Weight: weight,
			Value:  "Repository structure suggests complex multi-agent workflows",
		})
		totalScore += weight
	}

	// Blog/Tech Discussion Signals
	if strings.Contains(lowerDossier, "blog insight") {
		weight := 15
		if strings.Contains(lowerDossier, "evaluating move") || strings.Contains(lowerDossier, "scaling") {
			weight = 20
		}
		signals = append(signals, IntentSignal{
			Type:   "blog",
			Weight: weight,
			Value:  "Technical blog post analyzing orchestration constraints",
		})
		totalScore += weight
	}

	// Competitor Intelligence
	if strings.Contains(lowerDossier, "competitive intelligence") {
		weight := 0
		if strings.Contains(lowerDossier, "struggling") || strings.Contains(lowerDossier, "rate-limiting") {
			weight = 30 // High intent if they are failing with a competitor
			signals = append(signals, IntentSignal{
				Type:   "competitor",
				Weight: weight,
				Value:  "Evident friction with existing competitive solution",
			})
		} else if strings.Contains(lowerDossier, "evaluation phase") || strings.Contains(lowerDossier, "starred") {
			weight = 20 // Medium intent if they are just exploring
			signals = append(signals, IntentSignal{
				Type:   "competitor",
				Weight: weight,
				Value:  "Actively evaluating competitive frameworks",
			})
		}
		totalScore += weight
	}

	// Normalize Score (Cap at 100)
	if totalScore > 100 {
		totalScore = 100
	}

	return IntentScore{
		Score:   totalScore,
		Signals: signals,
	}
}
