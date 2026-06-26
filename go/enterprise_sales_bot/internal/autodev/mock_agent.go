package autodev

import (
	"context"
	"log/slog"
	"fmt"
)

// MockAgent is a simulated development agent for initial integration.
type MockAgent struct{}

func (m *MockAgent) ProposeSolution(ctx context.Context, task Task) (string, error) {
	slog.Info(fmt.Sprintf("MockAgent: Proposing solution for: %s", task.Description))
	return "Simulated changes", nil
}

func (m *MockAgent) ApplyChanges(ctx context.Context, proposal string) error {
	slog.Info(fmt.Sprintf("MockAgent: Applying changes: %s", proposal))
	return nil
}

func (m *MockAgent) Verify(ctx context.Context) error {
	slog.Info("MockAgent: Verifying changes...")
	return nil
}
