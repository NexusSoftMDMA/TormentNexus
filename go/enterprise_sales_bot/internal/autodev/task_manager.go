package autodev

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// TaskManager handles parsing of tasks from TODO.md and tracking their progress.
type TaskManager struct {
	todoPath string
}

// NewTaskManager creates a new TaskManager.
func NewTaskManager(todoPath string) *TaskManager {
	return &TaskManager{todoPath: todoPath}
}

// GetNextTask parses TODO.md and returns the highest priority uncompleted task
// whose dependencies have all been met.
func (m *TaskManager) GetNextTask(ctx context.Context) (*Task, error) {
	tasks, err := m.GetRunnableTasks(ctx)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	// Simple sort: High first
	for _, t := range tasks {
		if t.Category == "High" || strings.Contains(t.Description, "[HIGH]") {
			return &t, nil
		}
	}
	return &tasks[0], nil
}

// GetRunnableTasks parses TODO.md and returns ALL uncompleted tasks
// whose dependencies have all been met, suitable for concurrent execution.
func (m *TaskManager) GetRunnableTasks(ctx context.Context) ([]Task, error) {
	allTasks, err := m.ListAllTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Build map of completed tasks for fast lookup
	completedDesc := make(map[string]bool)
	var pendingTasks []Task

	for _, t := range allTasks {
		if t.Completed {
			completedDesc[t.Description] = true
		} else {
			pendingTasks = append(pendingTasks, t)
		}
	}

	if len(pendingTasks) == 0 {
		return nil, nil // All done!
	}

	var runnableTasks []Task
	for _, t := range pendingTasks {
		canRun := true
		for _, dep := range t.Dependencies {
			if !completedDesc[dep] {
				canRun = false
				break
			}
		}
		if canRun {
			runnableTasks = append(runnableTasks, t)
		}
	}

	if len(runnableTasks) == 0 {
		return nil, fmt.Errorf("deadlock: pending tasks exist but none have their dependencies met")
	}

	return runnableTasks, nil
}

// ListAllTasks returns all tasks from TODO.md.
func (m *TaskManager) ListAllTasks(ctx context.Context) ([]Task, error) {
	file, err := os.Open(m.todoPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var tasks []Task
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- [ ]") || strings.HasPrefix(line, "- [x]") {
			desc := strings.TrimSpace(line[5:])
			var deps []string

			// Extremely simple inline dependency parsing: "Task A (depends on: Task B, Task C)"
			if idx := strings.Index(desc, "(depends on:"); idx != -1 {
				depStr := strings.TrimSuffix(desc[idx+12:], ")")
				parts := strings.Split(depStr, ",")
				for _, p := range parts {
					deps = append(deps, strings.TrimSpace(p))
				}
				desc = strings.TrimSpace(desc[:idx])
			}

			tasks = append(tasks, Task{
				Description:  desc,
				Completed:    strings.HasPrefix(line, "- [x]"),
				Dependencies: deps,
			})
		}
	}
	return tasks, nil
}

// MarkCompleted updates TODO.md to mark a task as completed.
func (m *TaskManager) MarkCompleted(ctx context.Context, taskDescription string) error {
	// Simple implementation: read whole file, replace line, write back
	input, err := os.ReadFile(m.todoPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")
	found := false
	for i, line := range lines {
		// taskDescription from memory might not have the (depends on) suffix, so we do prefix matching
		if strings.Contains(line, "- [ ]") && strings.Contains(line, taskDescription) {
			lines[i] = strings.Replace(line, "- [ ]", "- [x]", 1)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task not found in TODO.md: %s", taskDescription)
	}

	output := strings.Join(lines, "\n")
	// #nosec G306 -- TODO file is intentionally world-readable
	err = os.WriteFile(m.todoPath, []byte(output), 0644) // #nosec G306
	// #nosec G703
	if err != nil {
		return err
	}

	return nil
}

// MarkTaskFailed marks a task as failed and resets its in-progress status
// so it can be retried in a future cycle (potentially with a different approach).
func (tm *TaskManager) MarkTaskFailed(id string) error {



	tasks, err := tm.ListAllTasks(context.Background())
	if err != nil {
		return err
	}

	for i := range tasks {
		if tasks[i].ID == id {
			// Do not mark as Complete. We just release it back to the pool
			// so another cycle can pick it up.
			// In a more robust system, we would increment a retry counter here.
			break
		}
	}

	return nil // State is inherently reset because we don't persist "in-progress" to disk yet
}
