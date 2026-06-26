package skillregistry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
)

// SkillEvolutionRecord tracks the performance of a specific skill version
type SkillEvolutionRecord struct {
	SkillID    string    `json:"skillId"`
	Version    int       `json:"version"`
	Successes  int       `json:"successes"`
	Failures   int       `json:"failures"`
	LastUsedAt time.Time `json:"lastUsedAt"`
}

func (r *SkillEvolutionRecord) WinRate() float64 {
	total := r.Successes + r.Failures
	if total == 0 {
		return 0
	}
	return float64(r.Successes) / float64(total)
}

// RecordOutcome updates the performance metrics for a skill
func (ds *SkillDecisionSystem) RecordOutcome(skillID string, success bool) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	id := strings.ToLower(skillID)
	skill, ok := ds.loaded[id]
	if !ok {
		return
	}

	skill.UseCount++
	skill.LastUsedAt = time.Now()

	// In a real impl, this would persist to a SkillEvolution table in SQLite
	fmt.Printf("[Evolution] Skill %s outcome recorded: success=%v\n", id, success)
}

// EvolveSkill uses an LLM to mutate a skill's content based on its performance history
func (ds *SkillDecisionSystem) EvolveSkill(ctx context.Context, skillID string, feedback string) error {
	skill, ok := ds.registry.Get(skillID)
	if !ok {
		return fmt.Errorf("skill %s not found", skillID)
	}

	prompt := fmt.Sprintf(`
		You are a Skill Evolution Engine.
		Task: Improve the following skill runbook based on user feedback.

		Skill Name: %s
		Current Description: %s
		Current Content:
		---
		%s
		---

		Feedback: %s

		Return the updated SKILL.md content only.
	`, skill.Name, skill.Description, skill.Content, feedback)

	resp, err := ai.AutoRoute(ctx, []ai.Message{
		{Role: "system", Content: "You are an expert prompt engineer and technical writer."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return err
	}

	// Update the skill in the registry
	skill.Content = resp.Content
	return ds.registry.Register(*skill)
}
