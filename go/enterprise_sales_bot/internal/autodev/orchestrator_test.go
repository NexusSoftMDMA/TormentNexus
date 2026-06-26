package autodev

import (
	"context"
	"os"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
)

func TestOrchestrator_ExecuteStep_Mock(t *testing.T) {
	// Setup mock environment
	tmpTodo, _ := os.CreateTemp("", "TODO_TEST.md")
	defer func() { _ = os.Remove(tmpTodo.Name()) }()
	if err := os.WriteFile(tmpTodo.Name(), []byte("- [ ] Test integration task"), 0644); err != nil {
		t.Fatalf("Failed to write mock TODO: %v", err)
	}

	manager := NewTaskManager(tmpTodo.Name())
	agent := &MockAgent{}
	prManager := &gitcheck.MockPRManager{}
	tracker := &deploy.MockCITracker{}

	// We use nil for DB for now, as ExecuteStep handles nil check for persistence
	// In a real integration test, we'd use a test DB.
	orchestrator := NewOrchestrator(nil, manager, agent, prManager, tracker)

	// Skip git sync logic for unit testing orchestrator flow
	_ = os.Setenv("SKIP_AUTODEV_SYNC", "true")
	_ = os.Setenv("GO_TEST_MODE", "true")
	_ = os.Setenv("SKIP_AUTODEV_TESTS", "true")
	defer func() { _ = os.Unsetenv("SKIP_AUTODEV_SYNC") }()
	defer func() { _ = os.Unsetenv("GO_TEST_MODE") }()
	defer func() { _ = os.Unsetenv("SKIP_AUTODEV_TESTS") }()

	ctx := context.Background()
	orchestrator.ExecuteStep(ctx)

	// Verify task was marked completed in mock file
	content, _ := os.ReadFile(tmpTodo.Name())
	if !contains(string(content), "- [x] Test integration task") {
		t.Error("Task was not marked completed in TODO file")
	}
}
