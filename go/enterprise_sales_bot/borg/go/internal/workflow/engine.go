package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// StepStatus represents the execution state of a workflow step
type StepStatus string

const (
	StatusPending   StepStatus = "pending"
	StatusRunning   StepStatus = "running"
	StatusCompleted StepStatus = "completed"
	StatusFailed    StepStatus = "failed"
	StatusSkipped   StepStatus = "skipped"
)

// Step represents a single unit of work in a workflow
type Step struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	DependsOn   []string          `json:"dependsOn,omitempty"`
	Status      StepStatus        `json:"status"`
	StartedAt   *time.Time        `json:"startedAt,omitempty"`
	FinishedAt  *time.Time        `json:"finishedAt,omitempty"`
	Error       string            `json:"error,omitempty"`
	Output      map[string]any    `json:"output,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`

	// Execute is the function that performs the step's work
	Execute func(ctx context.Context, inputs map[string]any) (map[string]any, error) `json:"-"`
}

// Workflow is a directed acyclic graph (DAG) of steps
type Workflow struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Steps       []*Step    `json:"steps"`
	Status      StepStatus `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	FinishedAt  *time.Time `json:"finishedAt,omitempty"`

	mu sync.RWMutex
}

// NewWorkflow creates a new workflow with the given steps
func NewWorkflow(id, name, description string, steps []*Step) *Workflow {
	for _, s := range steps {
		s.Status = StatusPending
	}
	return &Workflow{
		ID:          id,
		Name:        name,
		Description: description,
		Steps:       steps,
		Status:      StatusPending,
		CreatedAt:   time.Now().UTC(),
	}
}

// topologicalSort returns steps in execution order, respecting dependencies
func (w *Workflow) topologicalSort() ([]*Step, error) {
	stepMap := make(map[string]*Step, len(w.Steps))
	for _, s := range w.Steps {
		stepMap[s.ID] = s
	}

	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var sorted []*Step

	var visit func(id string) error
	visit = func(id string) error {
		if visited[id] {
			return nil
		}
		if visiting[id] {
			return fmt.Errorf("circular dependency detected at step %q", id)
		}

		step, ok := stepMap[id]
		if !ok {
			return fmt.Errorf("step %q not found", id)
		}

		visiting[id] = true
		for _, dep := range step.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[id] = false
		visited[id] = true
		sorted = append(sorted, step)
		return nil
	}

	for _, s := range w.Steps {
		if err := visit(s.ID); err != nil {
			return nil, err
		}
	}

	return sorted, nil
}

// Run executes all steps in topological order, running independent steps concurrently
func (w *Workflow) Run(ctx context.Context) error {
	w.mu.Lock()
	now := time.Now().UTC()
	w.StartedAt = &now
	w.Status = StatusRunning
	w.mu.Unlock()

	sorted, err := w.topologicalSort()
	if err != nil {
		w.mu.Lock()
		w.Status = StatusFailed
		fin := time.Now().UTC()
		w.FinishedAt = &fin
		w.mu.Unlock()
		return err
	}

	// Build outputs map for passing between steps
	outputs := &sync.Map{}

	// Build a completion tracker
	completed := make(map[string]chan struct{})
	for _, s := range sorted {
		completed[s.ID] = make(chan struct{})
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(sorted))

	for _, step := range sorted {
		wg.Add(1)
		go func(s *Step) {
			defer wg.Done()
			defer close(completed[s.ID])

			// Wait for all dependencies to complete
			for _, dep := range s.DependsOn {
				select {
				case <-ctx.Done():
					w.mu.Lock()
					s.Status = StatusSkipped
					w.mu.Unlock()
					return
				case <-completed[dep]:
					// Check if dependency failed
					w.mu.RLock()
					for _, ds := range w.Steps {
						if ds.ID == dep && ds.Status == StatusFailed {
							w.mu.RUnlock()
							w.mu.Lock()
							s.Status = StatusSkipped
							s.Error = fmt.Sprintf("dependency %q failed", dep)
							w.mu.Unlock()
							return
						}
					}
					w.mu.RUnlock()
				}
			}

			// Gather inputs from dependency outputs
			inputs := make(map[string]any)
			for _, dep := range s.DependsOn {
				if val, ok := outputs.Load(dep); ok {
					if m, ok := val.(map[string]any); ok {
						for k, v := range m {
							inputs[dep+"."+k] = v
						}
					}
				}
			}

			// Execute the step
			w.mu.Lock()
			startTime := time.Now().UTC()
			s.StartedAt = &startTime
			s.Status = StatusRunning
			w.mu.Unlock()

			output, execErr := s.Execute(ctx, inputs)

			w.mu.Lock()
			finTime := time.Now().UTC()
			s.FinishedAt = &finTime
			if execErr != nil {
				s.Status = StatusFailed
				s.Error = execErr.Error()
				errChan <- fmt.Errorf("step %q failed: %w", s.ID, execErr)
			} else {
				s.Status = StatusCompleted
				s.Output = output
				outputs.Store(s.ID, output)
			}
			w.mu.Unlock()
		}(step)
	}

	wg.Wait()
	close(errChan)

	w.mu.Lock()
	finTime := time.Now().UTC()
	w.FinishedAt = &finTime

	// Check for any failures
	var firstErr error
	for err := range errChan {
		if firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		w.Status = StatusFailed
	} else {
		w.Status = StatusCompleted
	}
	w.mu.Unlock()

	return firstErr
}

// Engine manages multiple workflows
type Engine struct {
	workflows map[string]*Workflow
	mu        sync.RWMutex
}

// NewEngine creates a new workflow engine
func NewEngine() *Engine {
	return &Engine{
		workflows: make(map[string]*Workflow),
	}
}

// Register adds a workflow to the engine
func (e *Engine) Register(wf *Workflow) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.workflows[wf.ID] = wf
}

// Get retrieves a workflow by ID
func (e *Engine) Get(id string) (*Workflow, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	wf, ok := e.workflows[id]
	return wf, ok
}

// List returns all registered workflows
func (e *Engine) List() []*Workflow {
	e.mu.RLock()
	defer e.mu.RUnlock()
	list := make([]*Workflow, 0, len(e.workflows))
	for _, wf := range e.workflows {
		list = append(list, wf)
	}
	return list
}

// RunWorkflow executes a workflow by ID
func (e *Engine) RunWorkflow(ctx context.Context, id string) error {
	wf, ok := e.Get(id)
	if !ok {
		return fmt.Errorf("workflow %q not found", id)
	}
	return wf.Run(ctx)
}
