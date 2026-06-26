package communication

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

// QualityScore represents the assessed quality of an outreach message.
type QualityScore struct {
	Overall		int		`json:"overall"`		// 0–100
	IsRelevant	bool		`json:"is_relevant"`		// Does it address the target's context?
	IsPersonalized	bool		`json:"is_personalized"`	// Uses company/contact-specific details?
	HasCTAsk	bool		`json:"has_cta"`
	Tone		string		`json:"tone"`		// professional, friendly, pushy, generic
	Issues		[]string	`json:"issues"`		// specific problems found
	Suggestions	[]string	`json:"suggestions"`	// improvement suggestions
}

// QualityScorer evaluates outreach quality before sending.
type QualityScorer struct {
	llmProvider	llm.LLMProvider	// optional LLM for semantic scoring
	minScore	int		// minimum acceptable score (0–100)
}

// NewQualityScorer creates a QualityScorer.
// minScore: messages scoring below this are flagged (default 60).
func NewQualityScorer(llmProvider llm.LLMProvider, minScore int) *QualityScorer {
	if minScore <= 0 {
		minScore = 60
	}
	return &QualityScorer{
		llmProvider:	llmProvider,
		minScore:	minScore,
	}
}

// Evaluate performs heuristic + LLM-assisted scoring of a message.
func (qs *QualityScorer) Evaluate(ctx context.Context, subject, body, contactName, companyName, role string) (QualityScore, error) {
	score := QualityScore{
		Overall:	50,	// start at neutral
		IsRelevant:	false,
		IsPersonalized:	false,
		HasCTAsk:	false,
		Tone:		"generic",
	}

	// 1. Check personalization signals
	if contactName != "" && (strings.Contains(body, contactName) || strings.Contains(subject, contactName)) {
		score.IsPersonalized = true
		score.Overall += 15
	} else {
		score.Issues = append(score.Issues, "No contact name found in message")
		score.Suggestions = append(score.Suggestions, "Include the recipient's name")
	}

	if companyName != "" && (strings.Contains(body, companyName) || strings.Contains(subject, companyName)) {
		score.IsPersonalized = true
		score.Overall += 10
	} else {
		score.Issues = append(score.Issues, "No company name referenced")
		score.Suggestions = append(score.Suggestions, "Mention the company by name")
	}

	if role != "" && strings.Contains(body, role) {
		score.IsPersonalized = true
		score.Overall += 5
	}

	// 2. Check relevance signals
	relevantKeywords := []string{
		"infrastructure", "orchestrat", "pipeline", "workflow",
		"automation", "efficiency", "scale", "platform",
		"tool", "integration", "deploy", "agent",
	}
	relevanceHits := 0
	for _, kw := range relevantKeywords {
		if strings.Contains(strings.ToLower(body), kw) {
			relevanceHits++
		}
	}
	if relevanceHits >= 2 {
		score.IsRelevant = true
		score.Overall += relevanceHits * 5
		if relevanceHits > 4 {
			score.Overall += 5	// bonus for very relevant
		}
	} else {
		score.Issues = append(score.Issues, "Message lacks relevant technical keywords")
		score.Suggestions = append(score.Suggestions, "Add context about the prospect's technical challenges")
	}

	// 3. Check for call-to-action
	ctaPhrases := []string{
		"let me know", "what do you think", "interested",
		"would you be open", "schedule", "can we", "tell me",
		"look forward", "hear your thoughts", "feedback",
	}
	for _, phrase := range ctaPhrases {
		if strings.Contains(strings.ToLower(body), phrase) {
			score.HasCTAsk = true
			score.Overall += 10
			break
		}
	}
	if !score.HasCTAsk {
		score.Issues = append(score.Issues, "No clear call to action")
		score.Suggestions = append(score.Suggestions, "Include a specific next step (reply/call/schedule)")
	}

	// 4. Tone analysis (heuristic)
	positiveWords := []string{"help", "solve", "opportunity", "value", "improve", "excited", "great", "thank"}
	pushyWords := []string{"must", "urgent", "limited", "act now", "don't miss", "exclusive"}
	positiveCount := 0
	pushyCount := 0
	lower := strings.ToLower(body)
	for _, w := range positiveWords {
		if strings.Contains(lower, w) {
			positiveCount++
		}
	}
	for _, w := range pushyWords {
		if strings.Contains(lower, w) {
			pushyCount++
		}
	}
	if pushyCount > 0 {
		score.Tone = "pushy"
		score.Overall -= pushyCount * 10
		score.Issues = append(score.Issues, "Message reads as pushy or urgent")
		score.Suggestions = append(score.Suggestions, "Use a consultative, value-first tone")
	} else if positiveCount >= 3 {
		score.Tone = "friendly"
		score.Overall += 5
	} else if positiveCount >= 1 {
		score.Tone = "professional"
		score.Overall += 3
	} else {
		score.Tone = "generic"
		score.Issues = append(score.Issues, "Message tone is generic")
		score.Suggestions = append(score.Suggestions, "Add positive, value-oriented language")
	}

	// 5. Length check
	wordCount := len(strings.Fields(body))
	if wordCount < 30 {
		score.Issues = append(score.Issues, "Message is too short (<30 words)")
		score.Suggestions = append(score.Suggestions, "Expand with relevant context about TormentNexus")
		score.Overall -= 10
	} else if wordCount > 300 {
		score.Issues = append(score.Issues, "Message is too long (>300 words)")
		score.Suggestions = append(score.Suggestions, "Shorten to keep attention")
		score.Overall -= 5
	} else {
		score.Overall += 5
	}

	// 6. Clamp score to 0–100
	if score.Overall < 0 {
		score.Overall = 0
	} else if score.Overall > 100 {
		score.Overall = 100
	}

	// 7. LLM-assisted refinement (if available)
	if qs.llmProvider != nil {
		refined, err := qs.llmRefine(ctx, subject, body, &score)
		if err != nil {
			slog.Info(fmt.Sprintf("QualityScorer: LLM refinement failed, using heuristic: %v", err))
		} else {
			score = *refined
		}
	}

	return score, nil
}

// llmRefine uses the LLM to provide additional quality assessment.
func (qs *QualityScorer) llmRefine(ctx context.Context, subject, body string, current *QualityScore) (*QualityScore, error) {
	prompt := fmt.Sprintf(`You are a quality assurance system for sales outreach. Evaluate the following sales email and return ONLY a JSON object with these fields:
{
  "overall": <0-100>,
  "is_relevant": <true/false>,
  "tone": "<professional|friendly|pushy|generic>",
  "issues": ["<issue1>", ...],
  "suggestions": ["<suggestion1>", ...]
}

Subject: %s
Body: %s

Return only valid JSON, no markdown.`, subject, body)

	resp, err := qs.llmProvider.Generate(ctx, llm.Prompt{
		User: prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM scoring failed: %w", err)
	}

	// Parse the response — for now just log it and return heuristic
	slog.Info(fmt.Sprintf("QualityScorer: LLM assessment: %s", resp))

	return current, nil
}

// IsPassing checks if the quality score meets the minimum threshold.
func (qs *QualityScorer) IsPassing(score QualityScore) bool {
	return score.Overall >= qs.minScore
}

// ScoreAndLog is a convenience method: evaluates, logs, and returns pass/fail.
func (qs *QualityScorer) ScoreAndLog(ctx context.Context, subject, body, contactName, companyName, role string) (QualityScore, bool) {
	score, err := qs.Evaluate(ctx, subject, body, contactName, companyName, role)
	if err != nil {
		slog.Info(fmt.Sprintf("QualityScorer: Score evaluation error: %v", err))
		return score, false
	}

	passing := qs.IsPassing(score)
	tag := "✅ PASS"
	if !passing {
		tag = "❌ FAIL"
	}

	slog.Info(fmt.Sprintf("QualityScorer [%s]: Score=%d/100 | Relevant=%v | Personalized=%v | CTA=%v | Tone=%s | Issues=%d | Suggests=%d",
		tag, score.Overall, score.IsRelevant, score.IsPersonalized, score.HasCTAsk, score.Tone, len(score.Issues), len(score.Suggestions)))

	if !passing {
		for _, issue := range score.Issues {
			slog.Info(fmt.Sprintf("  Issue: %s", issue))
		}
		for _, sugg := range score.Suggestions {
			slog.Info(fmt.Sprintf("  Suggestion: %s", sugg))
		}
	}

	return score, passing
}

// LogQualityMetrics logs aggregate quality metrics over time.
type QualityMetrics struct {
	TotalScored	int	`json:"total_scored"`
	Passed		int	`json:"passed"`
	Failed		int	`json:"failed"`
	AvgScore	float64	`json:"avg_score"`
	LastReset	time.Time
	TotalScoreSum	int
}
