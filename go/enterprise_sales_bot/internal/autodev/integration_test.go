package autodev

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
)

func TestAutonomousLoop_Integration(t *testing.T) {
	// Integration test for the orchestrator loop
	// Note: This requires a running PostgreSQL instance for db.DB

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database, err := db.NewDB(connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Create a temporary TODO.md
	content := "- [ ] Task Integration Test"
	tmpTodo, err := os.CreateTemp("", "TODO_integration.md")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpTodo.Name()) }()
	if err := os.WriteFile(tmpTodo.Name(), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write integration TODO: %v", err)
	}

	manager := NewTaskManager(tmpTodo.Name())
	agent := &MockAgent{}
	prManager := &gitcheck.MockPRManager{}
	tracker := &deploy.MockCITracker{}

	orchestrator := NewOrchestrator(database, manager, agent, prManager, tracker)

	// We skip the sync protocol to avoid git errors in CI environment
	_ = os.Setenv("SKIP_AUTODEV_SYNC", "true")
	defer func() { _ = os.Unsetenv("SKIP_AUTODEV_SYNC") }()

	// Run one cycle manually
	orchestrator.ExecuteStep(ctx)

	// Verify task is marked completed in the file
	newContent, _ := os.ReadFile(tmpTodo.Name())
	if !contains(string(newContent), "[x] Task Integration Test") {
		t.Errorf("Task was not marked completed in TODO.md. Content: %s", string(newContent))
	}
}

func TestMultiAgentWorkflow_Integration(t *testing.T) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		t.Skip("DATABASE_URL not set, skipping multi-agent integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	database, err := db.NewDB(connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Prepare multi-step TODO
	content := "- [ ] Step 1: Scrape Leads\n- [ ] Step 2: Enrich Contacts"
	tmpTodo, err := os.CreateTemp("", "TODO_multi.md")
	if err != nil {
		t.Fatalf("Failed to create multi TODO: %v", err)
	}
	defer func() { _ = os.Remove(tmpTodo.Name()) }()
	if err := os.WriteFile(tmpTodo.Name(), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write multi TODO: %v", err)
	}

	manager := NewTaskManager(tmpTodo.Name())
	agent := &MockAgent{}
	prManager := &gitcheck.MockPRManager{}
	tracker := &deploy.MockCITracker{}
	orchestrator := NewOrchestrator(database, manager, agent, prManager, tracker)

	_ = os.Setenv("SKIP_AUTODEV_SYNC", "true")
	defer func() { _ = os.Unsetenv("SKIP_AUTODEV_SYNC") }()

	// Execute two steps
	orchestrator.ExecuteStep(ctx)
	orchestrator.ExecuteStep(ctx)

	// Verify both completed
	newContent, _ := os.ReadFile(tmpTodo.Name())
	if !contains(string(newContent), "[x] Step 1") || !contains(string(newContent), "[x] Step 2") {
		t.Errorf("Workflow steps not completed. Content: %s", string(newContent))
	}
}

func TestLocalAgent_ApplyChanges(t *testing.T) {
	agent := &LocalAgent{}
	tmpFile := "test_apply.txt"
	defer func() { _ = os.Remove(tmpFile) }()

	proposal := fmt.Sprintf("FILE: %s\nCONTENT:\nHello World\n", tmpFile)
	err := agent.ApplyChanges(context.Background(), proposal)
	if err != nil {
		t.Fatalf("ApplyChanges failed: %v", err)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	if strings.TrimSpace(string(content)) != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", string(content))
	}
}

func TestLocalAgent_PathTraversalProtection(t *testing.T) {
	agent := &LocalAgent{}

	// Attempt to write outside the working directory
	proposal := "FILE: ../outside.txt\nCONTENT:\nMalicious Content\n"
	err := agent.ApplyChanges(context.Background(), proposal)

	if err == nil {
		t.Fatal("Expected error for path traversal attempt, got nil")
	}

	if !strings.Contains(err.Error(), "security: blocked attempt") {
		t.Errorf("Expected security block error, got: %v", err)
	}
}

func TestLocalAgent_Verify(t *testing.T) {
	agent := &LocalAgent{}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// We skip the tests inside Verify to avoid recursive test execution in this environment
	_ = os.Setenv("SKIP_AUTODEV_TESTS", "true")
	defer func() { _ = os.Unsetenv("SKIP_AUTODEV_TESTS") }()

	// This runs 'go build' which should pass in this environment
	err := agent.Verify(ctx)
	if err != nil {
		t.Errorf("LocalAgent.Verify failed: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr))
}
