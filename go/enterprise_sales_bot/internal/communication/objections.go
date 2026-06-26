// Package communication provides the objection handling library — a curated set of
// common B2B objections with proven counter-arguments, success-rate metadata,
// and context-aware matching.
package communication

import (
	"context"
	crypto_rand "crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func cryptoRandInt() int64 {
	n, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(1<<63-1))
	if err != nil {
		return 0 // Fallback to 0 if crypto fails to avoid panic
	}
	return n.Int64()
}


// ObjectionCategory groups objections by theme.
type ObjectionCategory string

const (
	CategoryPricing    ObjectionCategory = "pricing"
	CategorySecurity   ObjectionCategory = "security"
	CategoryTiming     ObjectionCategory = "timing"
	CategoryCompetition ObjectionCategory = "competition"
	CategoryNeed       ObjectionCategory = "need"
	CategoryAuthority  ObjectionCategory = "authority"
	CategoryVendorLockIn ObjectionCategory = "vendor_lock_in"
	CategoryMaturity   ObjectionCategory = "maturity"
	CategoryIntegration ObjectionCategory = "integration"
	CategorySupport    ObjectionCategory = "support"
)

// Objection describes a common B2B objection.
type Objection struct {
	ID        string            `json:"id"`
	Category  ObjectionCategory `json:"category"`
	Title     string            `json:"title"`
	Patterns  []string          `json:"patterns"`  // regex or keyword patterns
	Keywords  []string          `json:"keywords"`  // signal words
	Urgency   float64           `json:"urgency"`   // 0.0 – 1.0 (how critical to handle)
	Priority  int               `json:"priority"`  // higher = tried first
}

// ResponseOption is a single counter-argument with trackable metadata.
type ResponseOption struct {
	ID          string    `json:"id"`
	ObjectionID string    `json:"objection_id"`
	Text        string    `json:"text"`
	Approach    string    `json:"approach"`     // e.g. "value", "social-proof", "technical", "fear"
	UseCases    []string  `json:"use_cases"`    // deal stages or contexts where this works
	SuccessRate float64   `json:"success_rate"` // 0.0 – 1.0 (tracked over time)
	TimesUsed   int       `json:"times_used"`
	LastUsed    time.Time `json:"last_used"`
}

// ObjectionLibrary is the central registry of objections and responses.
type ObjectionLibrary struct {
	mu         sync.RWMutex
	objections []Objection
	responses  []ResponseOption
}

// NewObjectionLibrary creates a library populated with the embedded curated data.
func NewObjectionLibrary() *ObjectionLibrary {
	lib := &ObjectionLibrary{

	}
	if err := lib.loadEmbedded(); err != nil {
		slog.Warn("Failed to load embedded objection data, using empty library", "error", err)
	}
	return lib
}

// LoadJSON replaces the library contents with data from a JSON byte slice.
func (lib *ObjectionLibrary) LoadJSON(data []byte) error {
	var payload struct {
		Objections []Objection      `json:"objections"`
		Responses  []ResponseOption `json:"responses"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("parse objection data: %w", err)
	}
	lib.mu.Lock()
	defer lib.mu.Unlock()
	lib.objections = payload.Objections
	lib.responses = payload.Responses
	lib.assignDefaults()
	return nil
}

// MatchedResult holds the best objection and response for a given context.
type MatchedResult struct {
	Objection Objection
	Response  ResponseOption
	Score     float64 // 0.0 – 1.0 confidence of match
}

// MatchObjection finds the best objection and response for the given context.
// It returns nil if no good match is found.
func (lib *ObjectionLibrary) MatchObjection(ctx context.Context, text string, sentiment SentimentResult, dealStage db.LeadState) *MatchedResult {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	if len(lib.objections) == 0 || len(lib.responses) == 0 {
		return nil
	}

	textLower := strings.ToLower(text)

	var best *MatchedResult
	bestScore := 0.0

	for _, obj := range lib.objections {
		score := scoreObjection(obj, textLower, sentiment, dealStage)
		if score <= 0 {
			continue
		}
		resp := lib.bestResponseFor(obj.ID, dealStage)
		if resp == nil {
			continue
		}
		if score > bestScore {
			bestScore = score
			best = &MatchedResult{
				Objection: obj,
				Response:  *resp,
				Score:     score,
			}
		}
	}

	return best
}

// RecordOutcome updates the success-rate data for a response option.
func (lib *ObjectionLibrary) RecordOutcome(responseID string, success bool) {
	lib.mu.Lock()
	defer lib.mu.Unlock()

	for i := range lib.responses {
		if lib.responses[i].ID == responseID {
			lib.responses[i].TimesUsed++
			lib.responses[i].LastUsed = time.Now()
			if success {
				n := float64(lib.responses[i].TimesUsed)
				lib.responses[i].SuccessRate = ((lib.responses[i].SuccessRate * (n - 1)) + 1.0) / n
			} else {
				n := float64(lib.responses[i].TimesUsed)
				lib.responses[i].SuccessRate = (lib.responses[i].SuccessRate * (n - 1)) / n
			}
			break
		}
	}
}

// Statistics returns aggregate stats about the library.
func (lib *ObjectionLibrary) Statistics() map[string]interface{} {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	totalUsed := 0
	avgRate := 0.0
	for _, r := range lib.responses {
		totalUsed += r.TimesUsed
		avgRate += r.SuccessRate
	}
	if len(lib.responses) > 0 {
		avgRate /= float64(len(lib.responses))
	}

	return map[string]interface{}{
		"objection_count":   len(lib.objections),
		"response_count":    len(lib.responses),
		"total_times_used":  totalUsed,
		"average_success":   avgRate,
	}
}

// --- internal helpers ---

// loadEmbedded loads the compiled-in objection and response data.
func (lib *ObjectionLibrary) loadEmbedded() error {
	return lib.LoadJSON([]byte(embeddedObjectionData))
}

// assignDefaults fills any zero-value metadata after loading.
func (lib *ObjectionLibrary) assignDefaults() {
	for i := range lib.responses {
		if lib.responses[i].SuccessRate == 0 {
			lib.responses[i].SuccessRate = 0.5 // neutral start
		}
	}
}

// bestResponseFor selects the best response for an objection given the current deal stage.
func (lib *ObjectionLibrary) bestResponseFor(objectionID string, stage db.LeadState) *ResponseOption {
	var candidates []ResponseOption
	stageStr := strings.ToLower(string(stage))

	for _, resp := range lib.responses {
		if resp.ObjectionID != objectionID {
			continue
		}
		// Check if this response is appropriate for the stage
		if len(resp.UseCases) > 0 {
			applicable := false
			for _, uc := range resp.UseCases {
				if strings.EqualFold(uc, stageStr) || uc == "*" {
					applicable = true
					break
				}
			}
			if !applicable {
				continue
			}
		}
		candidates = append(candidates, resp)
	}

	if len(candidates) == 0 {
		// fallback — any response for this objection
		for _, resp := range lib.responses {
			if resp.ObjectionID == objectionID {
				candidates = append(candidates, resp)
			}
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Pick by weighted random (higher success rate = more likely)
	totalWeight := 0.0
	for _, c := range candidates {
		weight := c.SuccessRate * float64(1+c.TimesUsed) // prefer proven ones
		totalWeight += weight
	}
	r := float64(cryptoRandInt()) / float64(1<<63 - 1) * totalWeight
	accum := 0.0
	for _, c := range candidates {
		accum += c.SuccessRate * float64(1+c.TimesUsed)
		if accum >= r {
			return &c
		}
	}
	return &candidates[len(candidates)-1]
}

// scoreObjection computes how well an objection matches the given context.
func scoreObjection(obj Objection, textLower string, sentiment SentimentResult, _ db.LeadState) float64 {
	score := 0.0

	// 1. Keyword matches
	for _, kw := range obj.Keywords {
		if strings.Contains(textLower, strings.ToLower(kw)) {
			score += 0.4
		}
	}

	// 2. Pattern matches (simple substring match for now; could use regex)
	for _, pat := range obj.Patterns {
		if strings.Contains(textLower, strings.ToLower(pat)) {
			score += 0.5
		}
	}

	// 3. Sentiment boost
	if sentiment.Sentiment == SentimentNegative || sentiment.Sentiment == SentimentMixed {
		// Objections are more likely when sentiment is negative
		if sentiment.Confidence > 0.6 {
			score += 0.3 * sentiment.Confidence
		}
	}

	// 4. Urgency bonus
	if obj.Urgency > 0.7 {
		score *= 1.2
	}

	return score
}
