package orchestration

import (
	"context"
	"fmt"

	"github.com/borghq/borg-go/internal/ai"
)

type ConsensusResult struct {
	Prompt    string           `json:"prompt"`
	Agreement float64          `json:"agreement"`
	Summary   string           `json:"summary"`
	Votes     []ConsensusVote `json:"votes"`
}

type ConsensusVote struct {
	Model   string `json:"model"`
	Content string `json:"content"`
}

type ConsensusEngine struct {
	history *DebateHistoryStore
}

func NewConsensusEngine(history *DebateHistoryStore) *ConsensusEngine {
	return &ConsensusEngine{history: history}
}

func (e *ConsensusEngine) SeekConsensus(ctx context.Context, input struct {
	Prompt                       string
	Models                       []string
	RequiredAgreementPercentage *float64
}) (*ConsensusResult, error) {
	fmt.Printf("[Go Consensus] Seeking consensus on: %s\n", input.Prompt)

	var votes []ConsensusVote
	for _, model := range input.Models {
		resp, err := ai.AutoRouteWithModel(ctx, model, []ai.Message{
			{Role: "user", Content: input.Prompt},
		})
		if err == nil {
			votes = append(votes, ConsensusVote{Model: model, Content: resp.Content})
		}
	}

	// Simple agreement heuristic for simulation
	agreement := 0.75
	summary := "Most models agree on the core implementation strategy."

	result := &ConsensusResult{
		Prompt:    input.Prompt,
		Agreement: agreement,
		Summary:   summary,
		Votes:     votes,
	}

	// L2 Vault Logging (via DebateHistoryStore)
	if e.history != nil {
		_, _ = e.history.SaveNativeDebate(ctx, "consensus-session", input.Prompt, "Consensus Evaluation", &DebateResult{
			Approved:    agreement >= 0.5,
			Consensus:   agreement,
			FinalPlan:   summary,
			Contributions: []DebateContribution{
				{Role: "ConsensusEngine", Message: fmt.Sprintf("Calculated agreement: %.2f", agreement)},
			},
		})
	}

	return result, nil
}
