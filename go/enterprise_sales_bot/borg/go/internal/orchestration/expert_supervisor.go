package orchestration

import (
	"context"
	"fmt"
	"strings"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
)

type ExpertSupervisor struct {
	ModelID string
}

func NewExpertSupervisor(modelID string) *ExpertSupervisor {
	if modelID == "" {
		modelID = "gemini-2.5-flash"
	}
	return &ExpertSupervisor{ModelID: modelID}
}

type SupervisorCheckResult struct {
	IsComplete bool   `json:"isComplete"`
	Reason     string `json:"reason"`
	NextSteps  string `json:"nextSteps,omitempty"`
}

func (s *ExpertSupervisor) EvaluateProgress(ctx context.Context, goal string, transcript []string) (*SupervisorCheckResult, error) {
	prompt := fmt.Sprintf(`
		You are the TormentNexus Expert Supervisor.
		Your goal is to evaluate if the team has successfully completed the assigned task.
		
		ORIGINAL GOAL: %s
		
		CONVERSATION TRANSCRIPT:
		%s
		
		INSTRUCTIONS:
		1. Analyze the work performed so far.
		2. Determine if all requirements of the goal have been met.
		3. If complete, explain why and start your response with "COMPLETE".
		4. If NOT complete, list the remaining gaps and required next steps.
		
		Respond clearly and concisely.
	`, goal, strings.Join(transcript, "\n\n"))

	resp, err := ai.AutoRouteWithModel(ctx, s.ModelID, []ai.Message{
		{Role: "system", Content: "You are the TormentNexus Expert Supervisor."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(resp.Content)
	isComplete := strings.HasPrefix(strings.ToUpper(content), "COMPLETE")
	
	reason := content
	if isComplete {
		reason = strings.TrimSpace(strings.TrimPrefix(strings.ToUpper(content), "COMPLETE"))
		if strings.HasPrefix(reason, ":") {
			reason = strings.TrimSpace(strings.TrimPrefix(reason, ":"))
		}
	}

	return &SupervisorCheckResult{
		IsComplete: isComplete,
		Reason:     reason,
	}, nil
}
