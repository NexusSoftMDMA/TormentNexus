package orchestration

import (
	"context"
	"fmt"
	"strings"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
)

var autoRoute = ai.AutoRoute

type DebateResult struct {
	Approved      bool               `json:"approved"`
	FinalPlan     string             `json:"finalPlan"`
	Consensus     float64            `json:"consensus"`
	Contributions []DebateContribution `json:"contributions"`
}

type DebateContribution struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

// RunDebate executes a synchronous multi-agent council debate natively in Go.
// The Architect proposes an implementation strategy for the given objective.
// The Security Reviewer critiques it.
// The Lead Engineer synthesizes a final approved plan.
func RunDebate(ctx context.Context, history *DebateHistoryStore, objective string, contextData string) (*DebateResult, error) {
	contributions := []DebateContribution{}

	// 1. The Architect Proposes a Plan
	architectPrompt := fmt.Sprintf(`You are the Principal Architect.
Objective: %s
Context: %s

Propose a high-level implementation plan. Be concise and focus on structural integrity.`, objective, contextData)

	architectResp, err := autoRoute(ctx, []ai.Message{
		{Role: "user", Content: architectPrompt},
	})
	if err != nil {
		return nil, fmt.Errorf("architect failed: %w", err)
	}
	contributions = append(contributions, DebateContribution{
		Role:    "Architect",
		Message: architectResp.Content,
	})

	// 2. The Security Reviewer Critiques
	reviewerPrompt := fmt.Sprintf(`You are the Security Reviewer.
The Architect proposed this plan:
%s

Review this plan strictly for security vulnerabilities, edge cases, and failure modes. If it is safe, say "APPROVE". Otherwise, list concerns.`, architectResp.Content)

	reviewerResp, err := autoRoute(ctx, []ai.Message{
		{Role: "user", Content: reviewerPrompt},
	})
	if err != nil {
		return nil, fmt.Errorf("reviewer failed: %w", err)
	}
	contributions = append(contributions, DebateContribution{
		Role:    "Security Reviewer",
		Message: reviewerResp.Content,
	})

	// 3. The Lead Engineer Synthesizes and Approves
	leadPrompt := fmt.Sprintf(`You are the Lead Engineer.
Architect Plan:
%s

Reviewer Critique:
%s

Synthesize the final, actionable implementation plan incorporating the critique. If the reviewer rejected it fundamentally, output "REJECTED" as the first word.`, architectResp.Content, reviewerResp.Content)

	leadResp, err := autoRoute(ctx, []ai.Message{
		{Role: "user", Content: leadPrompt},
	})
	if err != nil {
		return nil, fmt.Errorf("lead engineer failed: %w", err)
	}
	contributions = append(contributions, DebateContribution{
		Role:    "Lead Engineer",
		Message: leadResp.Content,
	})

	approved := !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(leadResp.Content)), "REJECTED")

	// Simple heuristic consensus
	consensus := 1.0
	if !approved {
		consensus = 0.0
	} else if strings.Contains(strings.ToUpper(reviewerResp.Content), "APPROVE") {
		consensus = 1.0
	} else {
		consensus = 0.7 // Approved with modifications
	}

	result := &DebateResult{
		Approved:      approved,
		FinalPlan:     leadResp.Content,
		Consensus:     consensus,
		Contributions: contributions,
	}

	// L2 Vault Logging
	if history != nil {
		_, _ = history.SaveNativeDebate(ctx, "council-debate", objective, contextData, result)
	}

	return result, nil
}
