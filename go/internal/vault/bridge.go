package vault

import (
	"context"
	"fmt"
	"github.com/tormentnexushq/tormentnexus-go/internal/orchestration"
)

// RecordResolution saves a debate or consensus resolution to the L2 Vault.
func RecordResolution(ctx context.Context, history *orchestration.DebateHistoryStore, sessionID string, objective string, result *orchestration.DebateResult) error {
	if history == nil {
		return nil
	}
	_, err := history.SaveNativeDebate(ctx, sessionID, objective, "Vault Record", result)
	if err == nil {
		fmt.Printf("[Vault] Recorded resolution for mission: %s\n", sessionID)
	}
	return err
}
