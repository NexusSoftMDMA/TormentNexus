package memory

import (
	"context"
	"fmt"
	"time"
)

// L1Scratchpad represents ephemeral, fast memory tied to an active session
type L1Scratchpad struct {
	SessionID      string
	Prompt         string
	ToolOutputs    map[string]string
	ChainOfThought []string
	CreatedAt      time.Time
}

// MemoryEntry represents a single memory item to be persisted to L2
type MemoryEntry struct {
	ID        string
	SessionID string
	Type      string // "raw" or "heuristic"
	Content   string
	Timestamp time.Time
}

// L2Vault represents the permanent storage interface (e.g., SQLite vector DB)
type L2Vault interface {
	Commit(ctx context.Context, entry MemoryEntry) error
	SemanticSearch(ctx context.Context, query string, limit int) ([]MemoryEntry, error)
}

// MemoryBridge connects the L1 scratchpad to the L2 Vault
type MemoryBridge struct {
	vault L2Vault
}

func NewMemoryBridge(vault L2Vault) *MemoryBridge {
	return &MemoryBridge{vault: vault}
}

// Initialize populates a new L1 scratchpad with historical heuristics from L2
func (b *MemoryBridge) Initialize(ctx context.Context, sessionID string, initialPrompt string) (*L1Scratchpad, error) {
	fmt.Printf("[DualTier] Initializing L1 Scratchpad for session %s...\n", sessionID)

	// Query L2 for relevant historical context (favoring heuristics)
	results, err := b.vault.SemanticSearch(ctx, initialPrompt, 5)
	if err != nil {
		fmt.Printf("[DualTier] Warning: failed to fetch L2 context: %v\n", err)
	}

	cot := []string{}
	for _, res := range results {
		if res.Type == "heuristic" {
			cot = append(cot, fmt.Sprintf("Historical Lesson: %s", res.Content))
		}
	}

	return &L1Scratchpad{
		SessionID:      sessionID,
		Prompt:         initialPrompt,
		ToolOutputs:    make(map[string]string),
		ChainOfThought: cot,
		CreatedAt:      time.Now(),
	}, nil
}

// Teardown commits the L1 session to L2 and cleans up
func (b *MemoryBridge) Teardown(ctx context.Context, pad *L1Scratchpad, heuristicSummary string) error {
	fmt.Printf("[DualTier] Tearing down L1 Scratchpad and committing to L2 Vault...\n")

	// 1. Commit raw transcript
	rawContent := fmt.Sprintf("Prompt: %s\n", pad.Prompt)
	for k, v := range pad.ToolOutputs {
		rawContent += fmt.Sprintf("Tool %s Output: %s\n", k, v)
	}

	err := b.vault.Commit(ctx, MemoryEntry{
		ID:        fmt.Sprintf("raw-%s-%d", pad.SessionID, time.Now().Unix()),
		SessionID: pad.SessionID,
		Type:      "raw",
		Content:   rawContent,
		Timestamp: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to commit raw memory: %w", err)
	}

	// 2. Commit heuristic summary
	if heuristicSummary != "" {
		err = b.vault.Commit(ctx, MemoryEntry{
			ID:        fmt.Sprintf("heuristic-%s-%d", pad.SessionID, time.Now().Unix()),
			SessionID: pad.SessionID,
			Type:      "heuristic",
			Content:   heuristicSummary,
			Timestamp: time.Now(),
		})
		if err != nil {
			return fmt.Errorf("failed to commit heuristic memory: %w", err)
		}
	}

	return nil
}
