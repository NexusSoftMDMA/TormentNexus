package mcp

/**
 * @file predictor.go
 * @module go/internal/mcp
 *
 * WHAT: Go-native implementation of autonomous tool prediction.
 *
 * WHY: Total Autonomy — Go sidecar should be able to suggest relevant tools
 * independently of the Node control plane.
 */

import (
	"context"
	"fmt"
	"strings"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
)

type ToolPredictor struct {
	aggregator *Aggregator
}

func NewToolPredictor(agg *Aggregator) *ToolPredictor {
	return &ToolPredictor{aggregator: agg}
}

func (p *ToolPredictor) PredictAndPreload(ctx context.Context, chatHistory string, activeGoal string) ([]string, error) {
	topics, err := p.identifyTopics(ctx, chatHistory, activeGoal)
	if err != nil {
		return nil, err
	}

	var predictedTools []string
	for _, topic := range topics {
		// Use native ranking logic
		tools, _ := p.aggregator.ListTools(ctx)
		ranked := RankTools(topic, tools, 2)
		for _, r := range ranked {
			predictedTools = append(predictedTools, r.Name)
		}
	}

	return uniqueStrings(predictedTools), nil
}

func (p *ToolPredictor) identifyTopics(ctx context.Context, chatHistory string, activeGoal string) ([]string, error) {
	prompt := fmt.Sprintf(`
		You are the TormentNexus Supervisor Tool Predictor (Go Native).
		Analyze the following conversation and active goal. 
		Identify 3-5 specific technical topics or capabilities that the AI will likely need next.
		Focus on things that might require specialized MCP tools.

		GOAL: %s
		HISTORY: %s
		
		Return ONLY a comma-separated list of keywords.
	`, activeGoal, chatHistory)

	resp, err := ai.AutoRoute(ctx, []ai.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, err
	}

	parts := strings.Split(resp.Content, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(strings.ToLower(part))
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result, nil
}

func uniqueStrings(input []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, s := range input {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}
