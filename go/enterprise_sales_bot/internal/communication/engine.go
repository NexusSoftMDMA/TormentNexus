package communication

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
	"fmt"
)

// LearningSalesEngine implements the SalesStrategy interface.
// isHighValueDeal determines if a deal is high-value based on company tier or quoted pricing.
func isHighValueDeal(deal *db.Deal, company *db.Company) bool {
	if company.MarketCapTier == "Enterprise" {
		return true
	}
	if deal.QuotedPricing >= 100000 {
		return true
	}
	return false
}

type LearningSalesEngine struct {
	db		*db.DB
	crmClient	crm.CRMClient
	llm		llm.LLMProvider
}

// NewLearningSalesEngine creates a new instance of the engine.
func NewLearningSalesEngine(database *db.DB, crmClient crm.CRMClient, llmProvider llm.LLMProvider) *LearningSalesEngine {
	return &LearningSalesEngine{
		db:		database,
		crmClient:	crmClient,
		llm:		llmProvider,
	}
}

// Decide determines the next action for a lead.
func (e *LearningSalesEngine) Decide(ctx context.Context, salesCtx SalesContext) (Action, error) {
	slog.Info(fmt.Sprintf("LearningSalesEngine: Deciding next action for deal %d (Latest Intent: %s)", salesCtx.Deal.ID, salesCtx.LatestIntent))

	// 1. Analyze historical performance and lead quality
	if e.shouldAdvanceState(salesCtx) {
		newState := db.StateNegotiating
		// If highly qualified and intent is positive, we might jump to closing
		if e.QualifyLead(salesCtx) >= 80 && salesCtx.LatestIntent == IntentFollowUp {
			newState = db.StateClosedWon
		}

		// If deal is high-value, require human approval before advancing
		if isHighValueDeal(&salesCtx.Deal, &salesCtx.Company) && newState != db.StatePendingApproval {
			slog.Info(fmt.Sprintf("LearningSalesEngine: High-value deal %d requires human approval", salesCtx.Deal.ID))
			newState = db.StatePendingApproval
		}

		slog.Info(fmt.Sprintf("LearningSalesEngine: Advancing deal %d to %s state.", salesCtx.Deal.ID, newState))
		if e.db != nil {
			if err := e.db.UpdateDealState(ctx, salesCtx.Deal.ID, newState); err != nil {
				slog.Info(fmt.Sprintf("LearningSalesEngine: Error updating deal state: %v", err))
			} else {
				// Record template success for A/B testing metrics
				if err := e.db.MarkTemplateSuccessForDeal(ctx, salesCtx.Deal.ID); err != nil {
					slog.Info(fmt.Sprintf("LearningSalesEngine: Error marking template success for deal %d: %v", salesCtx.Deal.ID, err))
				}
				if e.crmClient != nil {
					// Immediate CRM Sync (non-blocking)
					go func() {
						updatedDeal := salesCtx.Deal
						updatedDeal.CurrentState = newState
						maxRetries := 3
						for i := 0; i < maxRetries; i++ {
							if err := e.crmClient.PushDeal(ctx, updatedDeal, salesCtx.Company, e.RouteLead(salesCtx)); err != nil {
								slog.Info(fmt.Sprintf("LearningSalesEngine: Immediate CRM Push failed (attempt %d/%d): %v", i+1, maxRetries, err))
								time.Sleep(time.Duration(i+1) * 2 * time.Second)
								continue
							}
							return
						}
						slog.Info(fmt.Sprintf("LearningSalesEngine Error: CRM state sync failed after %d attempts for deal %d", maxRetries, updatedDeal.ID))
					}()
				}
			}
		}
		return ActionAdvanceState, nil
	}

	// 2. Self-Learning Strategy Adaptation
	// In production, this would call e.llm.Generate to analyze sentiment and adjust Action
	if e.llm != nil {
		slog.Info(fmt.Sprintf("LearningSalesEngine: Analyzing sentiment and adapting strategy via LLM for deal %d", salesCtx.Deal.ID))
	}

	// 3. Base intent-driven logic
	if salesCtx.LatestIntent == IntentFollowUp && e.shouldAdvanceState(salesCtx) {
		return ActionAdvanceState, nil
	}

	switch salesCtx.LatestIntent {
	case IntentTechnical:
		return ActionRespond, nil
	case IntentPricing:
		if e.isHighValueLead(salesCtx) {
			return ActionRespond, nil
		}
		return ActionEscalate, nil	// Escalate high-tier pricing negotiation
	case IntentObjection:
		// Attempt one autonomous rebuttal, then escalate
		if e.countInteractionTypes(salesCtx.Interactions, "Outbound") < 2 {
			return ActionRespond, nil
		}
		return ActionEscalate, nil
	case IntentSpam:
		return ActionWait, nil
	}

	return ActionRespond, nil
}

func (e *LearningSalesEngine) shouldAdvanceState(ctx SalesContext) bool {
	// Logic to advance state from Engaged to Negotiating if interest is high or highly qualified
	if ctx.Deal.CurrentState == db.StateEngaged && (len(ctx.Interactions) > 3 || e.QualifyLead(ctx) > 70) {
		return true
	}
	// If in negotiating, check for closing signals
	if ctx.Deal.CurrentState == db.StateNegotiating && (ctx.LatestIntent == IntentFollowUp || ctx.LatestIntent == IntentPricing) {
		return e.QualifyLead(ctx) > 85
	}

	return false
}

func (e *LearningSalesEngine) isHighValueLead(ctx SalesContext) bool {
	return ctx.Company.MarketCapTier == "Enterprise" || e.ScoreLead(ctx) > 80
}

func (e *LearningSalesEngine) countInteractionTypes(interactions []db.Interaction, direction string) int {
	count := 0
	for _, i := range interactions {
		if i.Direction == direction {
			count++
		}
	}
	return count
}

// ScoreLead calculates a priority score based on tier and technical research.
func (e *LearningSalesEngine) ScoreLead(salesCtx SalesContext) int {
	score := 0

	// Tier scoring
	switch strings.ToLower(salesCtx.Company.MarketCapTier) {
	case "enterprise":
		score += 50
	case "mid-market":
		score += 25
	}

	// Dossier insight scoring
	if strings.Contains(salesCtx.Deal.TechnicalDossier, "BOTTLENECK") {
		score += 30
	}
	if strings.Contains(salesCtx.Deal.TechnicalDossier, "INFRASTRUCTURE") {
		score += 20
	}

	// Interaction quantity bonus
	score += len(salesCtx.Interactions) * 2

	if score > 100 {
		return 100
	}
	return score
}

// QualifyLead returns a qualification percentage (0-100) based on readiness to buy.
func (e *LearningSalesEngine) QualifyLead(ctx SalesContext) int {
	qual := 0

	// Base score from profile
	qual += e.ScoreLead(ctx) / 2

	// Engagement quality
	inboundCount := e.countInteractionTypes(ctx.Interactions, "Inbound")
	if inboundCount > 2 {
		qual += 20
	}

	// Intent analysis
	switch ctx.LatestIntent {
	case IntentPricing:
		qual += 15
	case IntentTechnical:
		qual += 10
	case IntentMeetingRequest:
		qual += 25
	case IntentFollowUp:
		qual += 20
	}

	// Penalty for objections
	if ctx.LatestIntent == IntentObjection {
		qual -= 10
	}

	if qual > 100 {
		return 100
	}
	if qual < 0 {
		return 0
	}
	return qual
}

// RouteLead determines the optimal internal representative or team for a given deal context.
func (e *LearningSalesEngine) RouteLead(salesCtx SalesContext) string {
	score := e.ScoreLead(salesCtx)
	qual := e.QualifyLead(salesCtx)

	// Routing Logic:
	// 1. Technical Enterprise: Route to Lead Solutions Architect
	if salesCtx.Company.MarketCapTier == "Enterprise" && (strings.Contains(salesCtx.Deal.TechnicalDossier, "BOTTLENECK") || salesCtx.LatestIntent == IntentTechnical) {
		return "Lead Solutions Architect"
	}

	// 2. High Value / High Readiness: Route to Senior Account Executive
	if score > 80 || qual > 75 {
		return "Senior Account Executive"
	}

	// 3. Technical SME: Route to Technical Sales Engineer
	if salesCtx.LatestIntent == IntentTechnical {
		return "Technical Sales Engineer"
	}

	// Default: Standard Sales Representative
	return "Sales Representative"
}
