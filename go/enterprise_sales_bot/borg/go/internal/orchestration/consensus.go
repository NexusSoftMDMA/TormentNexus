package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
)

type ConsensusEngine struct {
	history *DebateHistoryStore
	vault   controlplane.MemoryVault
}

func NewConsensusEngine(history *DebateHistoryStore, vault controlplane.MemoryVault) *ConsensusEngine {
	return &ConsensusEngine{
		history: history,
		vault:   vault,
	}
}

type ConsensusVote struct {
	Model   string  `json:"model"`
	Approved bool    `json:"approved"`
	Reason  string  `json:"reason"`
	Weight  float64 `json:"weight"`
}

type ConsensusOutcome struct {
	Agreed   bool            `json:"agreed"`
	Score    float64         `json:"score"`
	Summary  string          `json:"summary"`
	Votes    []ConsensusVote `json:"votes"`
}

func (e *ConsensusEngine) Resolve(ctx context.Context, task string, models []string) (*ConsensusOutcome, error) {
	fmt.Printf("[Go Consensus] Evaluating task: %s\n", task)

	var votes []ConsensusVote
	var totalWeight float64
	var weightedApproved float64

	for _, model := range models {
		// Weighted logic: Architect models get more weight
		weight := 1.0
		if strings.Contains(strings.ToLower(model), "sonnet") || strings.Contains(strings.ToLower(model), "gpt-4o") {
			weight = 2.0
		}

		resp, err := ai.AutoRouteWithModel(ctx, model, []ai.Message{
			{Role: "system", Content: "You are a Consensus Auditor. Review the task and plan. Respond with JSON: {\"approved\": boolean, \"reason\": \"string\"}"},
			{Role: "user", Content: task},
		})

		if err == nil {
			var voteResult struct {
				Approved bool   `json:"approved"`
				Reason   string `json:"reason"`
			}
			if err := json.Unmarshal([]byte(resp.Content), &voteResult); err == nil {
				vote := ConsensusVote{
					Model:    model,
					Approved: voteResult.Approved,
					Reason:   voteResult.Reason,
					Weight:   weight,
				}
				votes = append(votes, vote)
				totalWeight += weight
				if vote.Approved {
					weightedApproved += weight
				}
			}
		}
	}

	score := 0.0
	if totalWeight > 0 {
		score = weightedApproved / totalWeight
	}

	outcome := &ConsensusOutcome{
		Agreed:  score >= 0.66, // 2/3 weighted majority
		Score:   score,
		Summary: fmt.Sprintf("Consensus reached with score %.2f across %d models.", score, len(votes)),
		Votes:   votes,
	}

	// 1. Log to L2 Vault
	if e.vault != nil {
		fact := fmt.Sprintf("Consensus decision on task: %s. Outcome: %v, Score: %.2f", task, outcome.Agreed, outcome.Score)
		entry := controlplane.L2VaultRecord{
			ID:             fmt.Sprintf("consensus-%d", time.Now().UnixNano()),
			SessionID:      "system",
			Type:           controlplane.MemoryLongTerm,
			Content:        fact,
			Importance:     0.9,
			HeatScore:      100.0,
			LastAccessedAt: time.Now(),
			CreatedAt:      time.Now(),
		}
		_ = e.vault.Commit(ctx, entry)
	}

	// 2. Log to History
	if e.history != nil {
		_, _ = e.history.SaveNativeDebate(ctx, "consensus", task, outcome.Summary, &DebateResult{
			Approved: outcome.Agreed,
			Consensus: outcome.Score,
			FinalPlan: outcome.Summary,
		})
	}

	return outcome, nil
}

func (e *ConsensusEngine) SeekConsensus(ctx context.Context, input struct {
	Prompt                       string
	Models                       []string
	RequiredAgreementPercentage *float64
}) (*ConsensusOutcome, error) {
	return e.Resolve(ctx, input.Prompt, input.Models)
}
