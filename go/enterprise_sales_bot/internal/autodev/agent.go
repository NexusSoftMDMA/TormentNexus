package autodev

import "context"

// Task represents a unit of work for the autonomous development module.
type Task struct {
	Dependencies []string

	ID          string
	Description string
	Category    string
	Completed   bool
}

// Agent defines the interface for an autonomous developer.
type Agent interface {
	// ProposeSolution analyzes a task and proposes a set of changes.
	ProposeSolution(ctx context.Context, task Task) (string, error)
	// ApplyChanges applies the proposed solution to the codebase.
	ApplyChanges(ctx context.Context, proposal string) error
	// Verify runs tests and checks to ensure the changes are correct.
	Verify(ctx context.Context) error
}
