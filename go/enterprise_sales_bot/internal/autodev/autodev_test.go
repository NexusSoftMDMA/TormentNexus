package autodev

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestTaskManager_GetNextTask(t *testing.T) {
	content := `
# TODO
- [x] Task 1
- [ ] Task 2
- [ ] Task 3
`
	tmpfile, err := os.CreateTemp("", "TODO.md")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	manager := NewTaskManager(tmpfile.Name())
	task, err := manager.GetNextTask(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if task == nil {
		t.Fatal("Expected a task, got nil")
	}
	if task.Description != "Task 2" {
		t.Errorf("Expected 'Task 2', got '%s'", task.Description)
	}
}

func TestTaskManager_MarkCompleted(t *testing.T) {
	content := "- [ ] Task A\n- [ ] Task B"
	tmpfile, err := os.CreateTemp("", "TODO.md")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	if err := os.WriteFile(tmpfile.Name(), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write TODO file: %v", err)
	}

	manager := NewTaskManager(tmpfile.Name())
	err = manager.MarkCompleted(context.Background(), "Task A")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	newContent, _ := os.ReadFile(tmpfile.Name())
	if !strings.Contains(string(newContent), "- [x] Task A") {
		t.Errorf("Task A was not marked as completed. Content: %s", string(newContent))
	}
}
