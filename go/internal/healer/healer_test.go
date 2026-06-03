package healer

import (
	"context"
	"fmt"
	"testing"

	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
)

type mockVault struct {
	records []controlplane.L2VaultRecord
}

func (m *mockVault) Commit(ctx context.Context, entry controlplane.L2VaultRecord) error {
	m.records = append(m.records, entry)
	return nil
}

func (m *mockVault) SemanticSearch(ctx context.Context, query string, limit int) ([]controlplane.L2VaultRecord, error) {
	return nil, nil
}

func TestRecordHeal(t *testing.T) {
	vault := &mockVault{}
	hs := NewHealerService(nil, "", nil, vault)

	plan := FixPlan{
		ID: "test-fix",
		Diagnosis: Diagnosis{
			Description: "Test diagnosis",
		},
		Explanation: "Test explanation",
	}

	ctx := context.Background()
	hs.recordHeal(ctx, "test error", plan, true, 1)

	if len(vault.records) != 1 {
		t.Fatalf("expected 1 record in vault, got %d", len(vault.records))
	}

	record := vault.records[0]
	if record.SessionID != "kernel-healer" {
		t.Errorf("expected session ID kernel-healer, got %s", record.SessionID)
	}

	fmt.Printf("Vault Record: %s\n", record.Content)
}
