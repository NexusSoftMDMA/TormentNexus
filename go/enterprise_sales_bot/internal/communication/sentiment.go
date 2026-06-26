package communication

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

// Sentiment represents the emotional tone of a message.
type Sentiment int

const (
	SentimentUnknown	Sentiment	= iota
	SentimentPositive			// interested, enthusiastic, curious
	SentimentNeutral			// factual, non-committal
	SentimentNegative			// frustrated, dismissive, angry
	SentimentMixed				// contradictory signals
)

func (s Sentiment) String() string {
	switch s {
	case SentimentPositive:
		return "positive"
	case SentimentNeutral:
		return "neutral"
	case SentimentNegative:
		return "negative"
	case SentimentMixed:
		return "mixed"
	default:
		return "unknown"
	}
}

// SentimentResult wraps the analysis of a message.
type SentimentResult struct {
	Sentiment	Sentiment	`json:"sentiment"`
	Confidence	float64		`json:"confidence"`	// 0.0 – 1.0
	Score		int		`json:"score"`		// -100 (very negative) to +100 (very positive)
	Keywords	[]string	`json:"keywords"`	// detected signal words
	Urgency		float64		`json:"urgency"`	// 0.0 – 1.0 (how time-sensitive)
	Recommendation	string		`json:"recommendation"`	// suggested next action
}

// SentimentAnalyzer classifies inbound message sentiment.
type SentimentAnalyzer struct {
	llmProvider llm.LLMProvider	// optional LLM for semantic analysis
}

// NewSentimentAnalyzer creates a new analyzer.
func NewSentimentAnalyzer(llmProvider llm.LLMProvider) *SentimentAnalyzer {
	return &SentimentAnalyzer{llmProvider: llmProvider}
}

// Analyze performs heuristic + optional LLM sentiment analysis on a message.
func (sa *SentimentAnalyzer) Analyze(ctx context.Context, message string) (SentimentResult, error) {
	// Start with heuristic analysis
	result := sa.heuristicAnalysis(message)

	// If LLM is available, refine with semantic analysis
	if sa.llmProvider != nil {
		refined, err := sa.llmAnalysis(ctx, message, result)
		if err != nil {
			slog.Info(fmt.Sprintf("SentimentAnalyzer: LLM analysis failed, using heuristic: %v", err))
		} else {
			result = *refined
		}
	}

	return result, nil
}

// heuristicAnalysis uses keyword and pattern matching for fast classification.
func (sa *SentimentAnalyzer) heuristicAnalysis(message string) SentimentResult {
	lower := strings.ToLower(message)

	result := SentimentResult{
		Sentiment:	SentimentUnknown,
		Confidence:	0.5,
		Score:		0,
		Urgency:	0.0,
	}

	// Positive signal words
	positiveWords := []string{
		"interesting", "impressive", "great", "thanks", "thank you",
		"looks good", "tell me more", "curious", "let's", "schedule",
		"demo", "see it", "try it", "evaluate", "pilot", "yes",
		"sounds good", "excited", "potential", "value", "solution",
	}

	// Negative signal words
	negativeWords := []string{
		"not interested", "unsubscribe", "stop", "spam", "remove",
		"do not contact", "leave me alone", "no thanks", "not now",
		"too expensive", "budget", "waste", "annoying", "bad",
		"terrible", "useless", "irrelevant", "wrong",
	}

	// Urgency signal words
	urgencyWords := []string{
		"urgent", "critical", "asap", "immediately", "deadline",
		"emergency", "today", "now", "important", "pressing",
	}

	// Neutral signal patterns
	neutralPatterns := []string{
		"maybe", "not sure", "need to think", "consider",
		"forward to", "colleague", "team", "review",
		"information", "details", "specs", "website",
	}

	var positiveMatches, negativeMatches, urgencyMatches, neutralMatches []string

	for _, w := range positiveWords {
		if strings.Contains(lower, w) {
			positiveMatches = append(positiveMatches, w)
		}
	}
	for _, w := range negativeWords {
		if strings.Contains(lower, w) {
			negativeMatches = append(negativeMatches, w)
		}
	}
	for _, w := range urgencyWords {
		if strings.Contains(lower, w) {
			urgencyMatches = append(urgencyMatches, w)
		}
	}
	for _, w := range neutralPatterns {
		if strings.Contains(lower, w) {
			neutralMatches = append(neutralMatches, w)
		}
	}

	result.Keywords = append(result.Keywords, positiveMatches...)
	result.Keywords = append(result.Keywords, negativeMatches...)
	result.Keywords = append(result.Keywords, urgencyMatches...)
	result.Keywords = append(result.Keywords, neutralMatches...)

	// Score calculation
	posScore := len(positiveMatches) * 20
	negScore := len(negativeMatches) * 25
	neuScore := len(neutralMatches) * 10

	result.Score = posScore - negScore

	// Urgency
	if len(urgencyMatches) > 0 {
		result.Urgency = float64(len(urgencyMatches)) * 0.2
		if result.Urgency > 1.0 {
			result.Urgency = 1.0
		}
	}

	// Classification logic
	switch {
	case result.Score >= 40 && posScore > negScore*2:
		result.Sentiment = SentimentPositive
		result.Confidence = 0.7
	case result.Score <= -30 && negScore > posScore:
		result.Sentiment = SentimentNegative
		result.Confidence = 0.75
	case neuScore > posScore+negScore:
		result.Sentiment = SentimentNeutral
		result.Confidence = 0.6
	case posScore > 0 && negScore > 0:
		result.Sentiment = SentimentMixed
		result.Confidence = 0.5
	default:
		result.Sentiment = SentimentNeutral
		result.Confidence = 0.5
	}

	// Generate recommendation
	result.Recommendation = sa.generateRecommendation(result)

	return result
}

// llmAnalysis uses the LLM for deeper semantic sentiment analysis.
func (sa *SentimentAnalyzer) llmAnalysis(ctx context.Context, message string, heuristic SentimentResult) (*SentimentResult, error) {
	prompt := fmt.Sprintf(`Analyze the sentiment of this B2B sales email reply. Return ONLY a JSON object with these fields:
{
  "sentiment": "positive|negative|neutral|mixed",
  "confidence": <0.0-1.0>,
  "score": <-100 to 100>,
  "urgency": <0.0-1.0>,
  "recommendation": "<brief next-action suggestion>"
}

Message: %s

Return valid JSON only, no markdown.`, message)

	resp, err := sa.llmProvider.Generate(ctx, llm.Prompt{User: prompt})
	if err != nil {
		return nil, fmt.Errorf("LLM sentinel: %w", err)
	}

	// For now just log and return heuristic
	slog.Info(fmt.Sprintf("SentimentAnalyzer: LLM response: %s", resp))

	return &heuristic, nil
}

// generateRecommendation produces a suggested next action based on sentiment.
func (sa *SentimentAnalyzer) generateRecommendation(result SentimentResult) string {
	switch result.Sentiment {
	case SentimentPositive:
		if result.Urgency > 0.5 {
			return "High interest detected — escalate to senior rep or schedule demo immediately"
		}
		return "Positive signal — respond with additional technical detail and propose next step"
	case SentimentNegative:
		if result.Score < -60 {
			return "Strong negative — mark as lost and move to nurture sequence"
		}
		return "Negative signal — address concerns directly, offer clarification, or deprioritize"
	case SentimentNeutral:
		return "Neutral — provide additional value (case study, whitepaper, technical comparison)"
	case SentimentMixed:
		return "Mixed signals — clarify objections while reinforcing value proposition"
	default:
		return "Unknown sentiment — proceed with standard follow-up"
	}
}

// SummarizeDealSentiment aggregates sentiment across multiple interactions for a deal.
type DealSentimentSummary struct {
	DealID		int64	`json:"deal_id"`
	TotalMessages	int	`json:"total_messages"`
	AvgScore	float64	`json:"avg_score"`
	AvgUrgency	float64	`json:"avg_urgency"`
	MostCommon	string	`json:"most_common_sentiment"`
	Trend		string	`json:"trend"`	// improving, declining, stable
}

// AggregateDealSentiment combines multiple sentiment results into a summary.
func AggregateDealSentiment(results []SentimentResult) DealSentimentSummary {
	if len(results) == 0 {
		return DealSentimentSummary{}
	}

	var totalScore int
	var totalUrgency float64
	var posCount, negCount, neutralCount int

	for _, r := range results {
		totalScore += r.Score
		totalUrgency += r.Urgency
		switch r.Sentiment {
		case SentimentPositive:
			posCount++
		case SentimentNegative:
			negCount++
		default:
			neutralCount++
		}
	}

	n := len(results)
	summary := DealSentimentSummary{
		TotalMessages:	n,
		AvgScore:	float64(totalScore) / float64(n),
		AvgUrgency:	totalUrgency / float64(n),
	}

	// Most common sentiment
	if posCount > negCount && posCount > neutralCount {
		summary.MostCommon = "positive"
	} else if negCount > posCount && negCount > neutralCount {
		summary.MostCommon = "negative"
	} else {
		summary.MostCommon = "neutral"
	}

	// Trend (simple: compare first half to second half)
	if n >= 4 {
		mid := n / 2
		var firstHalfAvg, secondHalfAvg float64
		for i := 0; i < mid; i++ {
			firstHalfAvg += float64(results[i].Score)
		}
		for i := mid; i < n; i++ {
			secondHalfAvg += float64(results[i].Score)
		}
		firstHalfAvg /= float64(mid)
		secondHalfAvg /= float64(n - mid)

		diff := secondHalfAvg - firstHalfAvg
		if diff > 15 {
			summary.Trend = "improving"
		} else if diff < -15 {
			summary.Trend = "declining"
		} else {
			summary.Trend = "stable"
		}
	} else {
		summary.Trend = "insufficient data"
	}

	return summary
}

// AnalyzeSentiment performs a quick heuristic-based sentiment analysis on a text string.
// It does not require a SentimentAnalyzer instance or LLM — useful for simple checks.
func AnalyzeSentiment(text string) SentimentResult {
	if text == "" {
		return SentimentResult{
			Sentiment:  SentimentNeutral,
			Confidence: 0.5,
			Score:      0,
		}
	}

	textLower := strings.ToLower(text)
	words := strings.Fields(textLower)

	// Positive signal words
	positiveWords := map[string]bool{
		"great": true, "excellent": true, "interested": true, "perfect": true,
		"love": true, "amazing": true, "awesome": true, "yes": true,
		"thanks": true, "thank": true, "appreciate": true, "ready": true,
		"sure": true, "looks": true, "impressive": true, "good": true,
	}

	// Negative signal words
	negativeWords := map[string]bool{
		"bad": true, "terrible": true, "horrible": true, "hate": true,
		"expensive": true, "costly": true, "too": true, "won't": true,
		"cannot": true, "can't": true, "don't": true, "doesn't": true,
		"waste": true, "useless": true, "problem": true, "issue": true,
		"frustrated": true, "disappointed": true, "unfortunately": true,
		"wait": true, "busy": true, "not": true, "no": true,
	}

	// Objection signal phrases
	negativePhrases := []string{
		"too expensive", "over budget", "can't afford", "not interested",
		"not now", "wrong time", "too busy", "not for us", "doesn't work",
		"security concern", "privacy concern", "data privacy", "vendor lock",
		"not mature", "too complex", "not compatible",
	}

	posScore := 0
	negScore := 0

	for _, word := range words {
		if positiveWords[word] {
			posScore++
		}
		if negativeWords[word] {
			negScore++
		}
	}

	// Check multi-word phrases
	for _, phrase := range negativePhrases {
		if strings.Contains(textLower, phrase) {
			negScore += 2
		}
	}

	totalWords := len(words)
	if totalWords == 0 {
		totalWords = 1
	}

	score := 0
	if posScore > 0 || negScore > 0 {
		score = (posScore - negScore) * 100 / (posScore + negScore + totalWords)
	}

	var sentiment Sentiment
	switch {
	case score > 15:
		sentiment = SentimentPositive
	case score < -15:
		sentiment = SentimentNegative
	case score > 5:
		sentiment = SentimentPositive
	case score < -5:
		sentiment = SentimentNegative
	default:
		sentiment = SentimentNeutral
	}

	confidence := 0.5 + float64(posScore+negScore)/float64(totalWords+10)
	if confidence > 0.95 {
		confidence = 0.95
	}

	keywords := make([]string, 0)
	for _, word := range words {
		if positiveWords[word] || negativeWords[word] {
			keywords = append(keywords, word)
		}
	}

	return SentimentResult{
		Sentiment:  sentiment,
		Confidence: confidence,
		Score:      score,
		Keywords:   keywords,
	}
}
