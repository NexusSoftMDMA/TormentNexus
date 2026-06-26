package httpapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type localDarwinMutation struct {
	ID             string `json:"id"`
	OriginalPrompt string `json:"originalPrompt"`
	MutatedPrompt  string `json:"mutatedPrompt"`
	Reasoning      string `json:"reasoning"`
	Goal           string `json:"goal,omitempty"`
	Timestamp      int64  `json:"timestamp"`
}

type localDarwinExperiment struct {
	ID             string `json:"id"`
	MutationID     string `json:"mutationId"`
	Task           string `json:"task"`
	Status         string `json:"status"`
	ResultA        string `json:"resultA,omitempty"`
	ResultB        string `json:"resultB,omitempty"`
	Winner         string `json:"winner,omitempty"`
	JudgeReasoning string `json:"judgeReasoning,omitempty"`
	StartedAt      int64  `json:"startedAt"`
	CompletedAt    *int64 `json:"completedAt,omitempty"`
}

type persistedDarwinState struct {
	Mutations   []localDarwinMutation   `json:"mutations"`
	Experiments []localDarwinExperiment `json:"experiments"`
	NextID      int64                   `json:"nextId"`
}

type localDarwinStateManager struct {
	mu          sync.RWMutex
	persistPath string
	nextID      int64
	mutations   []localDarwinMutation
	experiments []localDarwinExperiment
}

func newLocalDarwinStateManager(persistPath string) *localDarwinStateManager {
	m := &localDarwinStateManager{
		persistPath: persistPath,
		nextID:      1,
		mutations:   []localDarwinMutation{},
		experiments: []localDarwinExperiment{},
	}
	m.load()
	return m
}

func (m *localDarwinStateManager) proposeMutation(originalPrompt string, goal string) localDarwinMutation {
	m.mu.Lock()
	defer m.mu.Unlock()
	mutationID := fmt.Sprintf("mut-%d", m.nextID)
	m.nextID++
	trimmedPrompt := strings.TrimSpace(originalPrompt)
	trimmedGoal := strings.TrimSpace(goal)
	mutatedPrompt := trimmedPrompt
	if mutatedPrompt == "" {
		mutatedPrompt = "[native-go-darwin-fallback] No original prompt was provided."
	}
	if trimmedGoal != "" {
		mutatedPrompt += "\n\n[Native Go Darwin fallback goal emphasis]\nPrioritize: " + trimmedGoal
		if len(mutatedPrompt) > 1200 {
			mutatedPrompt = mutatedPrompt[:1200]
		}
	}
	mutation := localDarwinMutation{
		ID:             mutationID,
		OriginalPrompt: originalPrompt,
		MutatedPrompt:  mutatedPrompt,
		Reasoning:      localDarwinReasoning(trimmedPrompt, trimmedGoal),
		Goal:           trimmedGoal,
		Timestamp:      time.Now().UTC().UnixMilli(),
	}
	m.mutations = append([]localDarwinMutation{mutation}, m.mutations...)
	m.saveLocked()
	return mutation
}

func (m *localDarwinStateManager) startExperiment(mutationID string, task string) (localDarwinExperiment, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	mutation, ok := m.findMutationLocked(mutationID)
	if !ok {
		return localDarwinExperiment{}, false
	}
	experimentID := fmt.Sprintf("exp-%d", m.nextID)
	m.nextID++
	experiment := localDarwinExperiment{
		ID:         experimentID,
		MutationID: mutationID,
		Task:       strings.TrimSpace(task),
		Status:     "PENDING",
		StartedAt:  time.Now().UTC().UnixMilli(),
	}
	m.experiments = append([]localDarwinExperiment{experiment}, m.experiments...)
	m.saveLocked()
	go m.runExperiment(experimentID, mutation)
	return experiment, true
}

func (m *localDarwinStateManager) runExperiment(experimentID string, mutation localDarwinMutation) {
	m.mu.Lock()
	index := m.experimentIndexLocked(experimentID)
	if index < 0 {
		m.mu.Unlock()
		return
	}
	m.experiments[index].Status = "RUNNING"
	task := m.experiments[index].Task
	m.saveLocked()
	m.mu.Unlock()

	resultA := localDarwinExecutionSummary("control", mutation.OriginalPrompt, task)
	resultB := localDarwinExecutionSummary("variant", mutation.MutatedPrompt, task)
	winner, reasoning := localDarwinVerdict(mutation, task)
	completedAt := time.Now().UTC().UnixMilli()

	m.mu.Lock()
	defer m.mu.Unlock()
	index = m.experimentIndexLocked(experimentID)
	if index < 0 {
		return
	}
	m.experiments[index].ResultA = resultA
	m.experiments[index].ResultB = resultB
	m.experiments[index].Winner = winner
	m.experiments[index].JudgeReasoning = reasoning
	m.experiments[index].Status = "COMPLETED"
	m.experiments[index].CompletedAt = &completedAt
	m.saveLocked()
}

func (m *localDarwinStateManager) status() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mutations := make([]localDarwinMutation, len(m.mutations))
	copy(mutations, m.mutations)
	experiments := make([]localDarwinExperiment, len(m.experiments))
	copy(experiments, m.experiments)
	running := false
	for _, experiment := range experiments {
		if experiment.Status == "RUNNING" || experiment.Status == "PENDING" {
			running = true
			break
		}
	}
	return map[string]any{
		"running":         running,
		"mutationCount":   len(mutations),
		"experimentCount": len(experiments),
		"mutations":       mutations,
		"experiments":     experiments,
	}
}

func (m *localDarwinStateManager) load() {
	if strings.TrimSpace(m.persistPath) == "" {
		return
	}
	data, err := os.ReadFile(m.persistPath)
	if err != nil {
		return
	}
	var state persistedDarwinState
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}
	m.mutations = state.Mutations
	m.experiments = state.Experiments
	if state.NextID > 0 {
		m.nextID = state.NextID
	}
}

func (m *localDarwinStateManager) saveLocked() {
	if strings.TrimSpace(m.persistPath) == "" {
		return
	}
	state := persistedDarwinState{
		Mutations:   append([]localDarwinMutation(nil), m.mutations...),
		Experiments: append([]localDarwinExperiment(nil), m.experiments...),
		NextID:      m.nextID,
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(m.persistPath), 0o755)
	_ = os.WriteFile(m.persistPath, data, 0o644)
}

func (m *localDarwinStateManager) findMutationLocked(id string) (localDarwinMutation, bool) {
	for _, mutation := range m.mutations {
		if mutation.ID == id {
			return mutation, true
		}
	}
	return localDarwinMutation{}, false
}

func (m *localDarwinStateManager) experimentIndexLocked(id string) int {
	for i, experiment := range m.experiments {
		if experiment.ID == id {
			return i
		}
	}
	return -1
}

func localDarwinReasoning(originalPrompt string, goal string) string {
	parts := []string{"Native Go Darwin fallback generated a deterministic local mutation scaffold instead of a live LLM-authored mutation."}
	if goal != "" {
		parts = append(parts, "It emphasized the requested goal: "+goal+".")
	}
	if len(strings.Fields(originalPrompt)) > 40 {
		parts = append(parts, "The original prompt was long enough that a concise goal emphasis is likely easier to compare during local fallback experiments.")
	}
	return strings.Join(parts, " ")
}

func localDarwinExecutionSummary(mode string, prompt string, task string) string {
	trimmedPrompt := strings.TrimSpace(prompt)
	if len(trimmedPrompt) > 220 {
		trimmedPrompt = trimmedPrompt[:220] + "..."
	}
	return fmt.Sprintf("[Native Go Darwin fallback %s execution]\nTask: %s\nPrompt Snapshot: %s", strings.ToUpper(mode), strings.TrimSpace(task), trimmedPrompt)
}

func localDarwinVerdict(mutation localDarwinMutation, task string) (string, string) {
	originalLen := len(strings.Fields(strings.TrimSpace(mutation.OriginalPrompt)))
	mutatedLen := len(strings.Fields(strings.TrimSpace(mutation.MutatedPrompt)))
	if strings.TrimSpace(task) == "" {
		return "TIE", "Native Go Darwin fallback had no task-specific benchmark input, so it returned a tie instead of over-claiming a winner."
	}
	if mutatedLen > 0 && mutatedLen < originalLen {
		return "B", "Native Go Darwin fallback picked the variant because the mutated prompt is more concise while still preserving the requested goal emphasis."
	}
	if originalLen > 0 && originalLen < mutatedLen {
		return "A", "Native Go Darwin fallback kept the control prompt because the variant became longer without a stronger local heuristic signal."
	}
	return "TIE", "Native Go Darwin fallback could not derive a stronger deterministic winner signal from local prompt-length heuristics, so it returned a tie."
}
