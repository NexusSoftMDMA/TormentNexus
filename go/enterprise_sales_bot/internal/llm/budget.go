package llm

import (
	"context"
	"sync"
	"time"
)

// TokenBudget tracks LLM token usage against a configurable budget.
type TokenBudget struct {
	BudgetTokens    int64         `json:"budget_tokens"`     // total allowed tokens
	UsedTokens      int64         `json:"used_tokens"`       // tokens consumed
	ResetInterval   time.Duration `json:"reset_interval"`    // e.g., 24h, 7d
	LastReset       time.Time     `json:"last_reset"`        // when counter was last reset
	WarningThreshold float64      `json:"warning_threshold"` // 0.8 = warn at 80% usage
	OnBudgetExceeded func()        `json:"-"`                 // optional callback when exceeded
	mu              sync.RWMutex
}

// TokenUsage records a single LLM call's token consumption.
type TokenUsage struct {
	Tokens      int       `json:"tokens"`
	Model       string    `json:"model"`
	Prompt      string    `json:"prompt"`
	Response    string    `json:"response"`
	CostUSD     float64   `json:"cost_usd"` // optional: track actual cost
	Timestamp   time.Time `json:"timestamp"`
	DealID      int64     `json:"deal_id,omitempty"`     // optional: attribute to deal
	ContactID   int64     `json:"contact_id,omitempty"`  // optional: attribute to contact
	Purpose     string    `json:"purpose"`               // e.g., "intent_classification", "response_generation"
}

// NewTokenBudget creates a TokenBudget with the specified limit and reset interval.
// onExceeded is called when usage exceeds budget (can be nil).
func NewTokenBudget(budgetTokens int64, resetInterval time.Duration, warningThreshold float64, onExceeded func()) *TokenBudget {
	return &TokenBudget{
		BudgetTokens:       budgetTokens,
		UsedTokens:         0,
		ResetInterval:      resetInterval,
		LastReset:          time.Now(),
		WarningThreshold:   warningThreshold,
		OnBudgetExceeded:   onExceeded,
	}
}

// DefaultTokenBudget returns a sensible default budget (1M tokens/day, warn at 80%).
func DefaultTokenBudget() *TokenBudget {
	return NewTokenBudget(1_000_000, 24*time.Hour, 0.8, nil)
}

// RecordUsage logs token consumption and checks budget limits.
// Returns true if within budget, false if exceeded.
func (tb *TokenBudget) RecordUsage(tokens int, dealID, contactID int64, model, prompt, response, purpose string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Check if we need to reset
	if time.Since(tb.LastReset) > tb.ResetInterval {
		tb.UsedTokens = 0
		tb.LastReset = time.Now()
	}

	tb.UsedTokens += int64(tokens)

	// Check if exceeded
	exceeded := tb.UsedTokens >= tb.BudgetTokens
	if exceeded && tb.OnBudgetExceeded != nil {
		go tb.OnBudgetExceeded() // call callback in background
	}

	return !exceeded
}

// IsWithinBudget checks current usage without recording anything.
func (tb *TokenBudget) IsWithinBudget() bool {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	// Check if we need to reset
	if time.Since(tb.LastReset) > tb.ResetInterval {
		return true // would reset on next use
	}

	return tb.UsedTokens < tb.BudgetTokens
}

// GetUsage returns current usage stats.
func (tb *TokenBudget) GetUsage() (used int64, budget int64, remaining int64, percentUsed float64) {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	// Check if we need to reset
	if time.Since(tb.LastReset) > tb.ResetInterval {
		return 0, tb.BudgetTokens, tb.BudgetTokens, 0.0
	}

	used = tb.UsedTokens
	budget = tb.BudgetTokens
	remaining = tb.BudgetTokens - used
	if remaining < 0 {
		remaining = 0
	}
	percentUsed = float64(used) / float64(tb.BudgetTokens)
	return
}

// ShouldWarn checks if usage has exceeded the warning threshold.
func (tb *TokenBudget) ShouldWarn() bool {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	if time.Since(tb.LastReset) > tb.ResetInterval {
		return false
	}

	percentUsed := float64(tb.UsedTokens) / float64(tb.BudgetTokens)
	return percentUsed >= tb.WarningThreshold
}

// EstimatedCost returns estimated cost based on tokens used.
// modelCostPer1K: cost per 1000 tokens for the model (e.g., 0.002 for $0.002/1K tokens)
func (tb *TokenBudget) EstimatedCost(modelCostPer1K float64) float64 {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	return float64(tb.UsedTokens) * modelCostPer1K / 1000.0
}

// DealTokenTracker tracks token usage per deal for granular budgeting.
type DealTokenTracker struct {
	dealUsage map[int64]*TokenUsage  // dealID -> latest usage
	mu        sync.RWMutex
}

// NewDealTokenTracker creates a new deal-level token tracker.
func NewDealTokenTracker() *DealTokenTracker {
	return &DealTokenTracker{
		dealUsage: make(map[int64]*TokenUsage),
	}
}

// RecordUsage logs usage for a specific deal.
func (dtt *DealTokenTracker) RecordUsage(dealID int64, tokens int, model, prompt, response, purpose string) {
	dtt.mu.Lock()
	defer dtt.mu.Unlock()

	dtt.dealUsage[dealID] = &TokenUsage{
		Tokens:    tokens,
		Model:     model,
		Prompt:    prompt,
		Response:  response,
		Timestamp: time.Now(),
		DealID:    dealID,
		Purpose:   purpose,
	}
}

// GetDealUsage retrieves the latest usage for a deal.
func (dtt *DealTokenTracker) GetDealUsage(dealID int64) *TokenUsage {
	dtt.mu.RLock()
	defer dtt.mu.RUnlock()

	return dtt.dealUsage[dealID]
}

// GetAllUsage returns all deal usage records.
func (dtt *DealTokenTracker) GetAllUsage() []TokenUsage {
	dtt.mu.RLock()
	defer dtt.mu.RUnlock()

	usages := make([]TokenUsage, 0, len(dtt.dealUsage))
	for _, usage := range dtt.dealUsage {
		usages = append(usages, *usage)
	}
	return usages
}

// GetTotalTokens sums all tokens used across deals.
func (dtt *DealTokenTracker) GetTotalTokens() int64 {
	dtt.mu.RLock()
	defer dtt.mu.RUnlock()

	var total int64
	for _, usage := range dtt.dealUsage {
		total += int64(usage.Tokens)
	}
	return total
}

// BudgetAwareProvider wraps an LLMProvider with budget tracking.
type BudgetAwareProvider struct {
	provider    LLMProvider
	budget      *TokenBudget
	dealTracker *DealTokenTracker
	modelCostPer1K float64 // cost per 1000 tokens for pricing
}

// NewBudgetAwareProvider wraps a provider with budget tracking.
func NewBudgetAwareProvider(provider LLMProvider, budget *TokenBudget, dealTracker *DealTokenTracker, modelCostPer1K float64) *BudgetAwareProvider {
	return &BudgetAwareProvider{
		provider:       provider,
		budget:         budget,
		dealTracker:    dealTracker,
		modelCostPer1K: modelCostPer1K,
	}
}

// Generate wraps the provider's Generate with budget tracking.
func (bap *BudgetAwareProvider) Generate(ctx context.Context, prompt Prompt) (string, error) {
	// Check budget before calling
	if !bap.budget.IsWithinBudget() {
		return "", ErrBudgetExceeded
	}

	// Call underlying provider
	response, err := bap.provider.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}

	// Estimate tokens (simple heuristic: ~4 chars per token)
	estimatedTokens := (len(prompt.System) + len(prompt.User) + len(response)) / 4

	// Record usage
	bap.budget.RecordUsage(estimatedTokens, 0, 0, "", prompt.User, response, "generate")

	// Warn if approaching limit
	if bap.budget.ShouldWarn() {
		// Log warning (in production, this could send alerts)
		used, _, _, _ := bap.budget.GetUsage()
		_ = used
	}

	return response, nil
}

// ErrBudgetExceeded is returned when the token budget is exhausted.
var ErrBudgetExceeded = &BudgetError{message: "token budget exceeded"}

// BudgetError represents a budget-related error.
type BudgetError struct {
	message string
}

func (e *BudgetError) Error() string {
	return e.message
}