package skillregistry

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
)

// EvolveSystemPrompt reviews skill outcomes and decays, prompting the LLM to refine the agent persona.
func EvolveSystemPrompt(ctx context.Context, db *sql.DB) error {
	lowSkills, err := GetLowPerformingSkills(ctx, db, 0.6)
	if err != nil {
		return fmt.Errorf("evolve prompt: get low performing skills: %w", err)
	}

	var currentPersona string
	_ = db.QueryRowContext(ctx, `SELECT value FROM core_memory_scratchpad WHERE key = 'persona'`).Scan(&currentPersona)
	if currentPersona == "" {
		currentPersona = "You are a helpful coding assistant."
	}

	if len(lowSkills) == 0 {
		return nil
	}

	skillSummary := ""
	for _, s := range lowSkills {
		skillSummary += fmt.Sprintf("- Skill: %s, Success: %d, Failure: %d, Win-Rate: %.2f%%\n",
			s.SkillName, s.SuccessCount, s.FailureCount, s.WinRate*100)
	}

	userPrompt := fmt.Sprintf(`Our agent has encountered execution difficulties. Analyze the following low-performing skills/tools and refine the current system persona to prevent future failures when using these tools.

Low-Performing Skills telemetry:
%s
Current System Persona:
---
%s
---

Task: Output the newly updated system persona. It must include explicit instructions, behavioral constraints, and tips on how to correctly call/use the troubled tools. Output ONLY the new persona text. Do not include markdown code block quotes or conversational greetings.`, skillSummary, currentPersona)

	messages := []ai.Message{
		{Role: "system", Content: "You are an advanced agent engineering optimizer. Output the refined persona text directly."},
		{Role: "user", Content: userPrompt},
	}

	resp, errRoute := ai.AutoRoute(ctx, messages)
	if errRoute != nil {
		return fmt.Errorf("evolve prompt LLM call: %w", errRoute)
	}

	refined := resp.Content
	if refined == "" {
		return fmt.Errorf("evolve prompt: LLM returned empty response")
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO core_memory_scratchpad (key, value, updated_at)
		VALUES ('persona', ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at
	`, refined, time.Now().UTC().Format("2006-01-02 15:04:05"))
	if err != nil {
		return fmt.Errorf("evolve prompt: save refined persona: %w", err)
	}

	return nil
}

// StartPromptEvolutionLoop runs the prompt evolution routine periodically.
func StartPromptEvolutionLoop(ctx context.Context, db *sql.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = EvolveSystemPrompt(ctx, db)
			}
		}
	}()
}
