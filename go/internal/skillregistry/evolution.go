package skillregistry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
)

// SkillEvolutionRecord tracks the performance of a specific skill version
// Evolution Config
const (
	MinUsesForRetirement = 5
	RetirementThreshold  = 0.3 // < 30% win rate gets retired
)

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

	// Update evolution metrics directly on the loaded skill struct assuming we add these fields
	if success {
		skill.Successes++
	} else {
		skill.Failures++
	}

	// Auto-retirement check
	total := skill.Successes + skill.Failures
	if total >= MinUsesForRetirement {
		winRate := float64(skill.Successes) / float64(total)
		if winRate < RetirementThreshold {
			fmt.Printf("[Evolution] Auto-retiring skill %s due to low win rate (%.2f)\n", id, winRate)
			// Retire the skill
			skill.IsRetired = true
			ds.registry.Unregister(skill.Name)
			delete(ds.loaded, id)
			return
		}
	}

	// In a real impl, this would persist to a SkillEvolution table in SQLite
	fmt.Printf("[Evolution] Skill %s outcome recorded: success=%v\n", id, success)
}

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
