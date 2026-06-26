// Package session provides session lifecycle management ported from
// packages/core/src/services/SessionManager.ts.
//
// It tracks active sessions, their states, and provides CRUD operations.
package session

import (
	"fmt"
	"sync"
	"time"
)

// SessionState represents the lifecycle state of a session.
type SessionState string

const (
	StateCreated  SessionState = "created"
	StateStarting SessionState = "starting"
	StateRunning  SessionState = "running"
	StateStopping SessionState = "stopping"
	StateStopped  SessionState = "stopped"
	StateFailed   SessionState = "failed"
	StatePaused   SessionState = "paused"
)

// Session represents a tracked session.
type Session struct {
	ID          string            `json:"id"`
	CLIType     string            `json:"cliType"`
	State       SessionState      `json:"state"`
	Task        string            `json:"task,omitempty"`
	WorkDir     string            `json:"workDir,omitempty"`
	StartedAt   int64             `json:"startedAt,omitempty"`
	StoppedAt   int64             `json:"stoppedAt,omitempty"`
	SourcePath  string            `json:"sourcePath,omitempty"`
	Format      string            `json:"sessionFormat,omitempty"`
	Valid       bool              `json:"valid"`
	Models      []string          `json:"detectedModels,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SessionManager manages session lifecycle.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	maxSize  int
	onChange func(session *Session, oldState SessionState)
}

// NewSessionManager creates a new session manager.
func NewSessionManager(maxSize int) *SessionManager {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &SessionManager{
		sessions: make(map[string]*Session),
		maxSize:  maxSize,
	}
}

// OnChange registers a callback for session state changes.
func (sm *SessionManager) OnChange(fn func(session *Session, oldState SessionState)) {
	sm.onChange = fn
}

// Create creates a new session.
func (sm *SessionManager) Create(id, cliType, workDir, task string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := &Session{
		ID:        id,
		CLIType:   cliType,
		State:     StateCreated,
		Task:      task,
		WorkDir:   workDir,
		StartedAt: time.Now().UnixMilli(),
		Valid:     true,
		Models:    []string{},
		Metadata:  make(map[string]string),
	}

	sm.sessions[id] = session
	return session
}

// Start transitions a session to running state.
func (sm *SessionManager) Start(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s, ok := sm.sessions[id]
	if !ok {
		return fmt.Errorf("session %s not found", id)
	}

	oldState := s.State
	s.State = StateRunning
	s.StartedAt = time.Now().UnixMilli()

	if sm.onChange != nil {
		sm.onChange(s, oldState)
	}
	return nil
}

// Stop transitions a session to stopped state.
func (sm *SessionManager) Stop(id string) error {
	return sm.transition(id, StateStopped)
}

// Pause transitions a session to paused state.
func (sm *SessionManager) Pause(id string) error {
	return sm.transition(id, StatePaused)
}

// Resume transitions a paused session back to running.
func (sm *SessionManager) Resume(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s, ok := sm.sessions[id]
	if !ok {
		return fmt.Errorf("session %s not found", id)
	}
	if s.State != StatePaused {
		return fmt.Errorf("session %s is not paused (state: %s)", id, s.State)
	}

	oldState := s.State
	s.State = StateRunning
	if sm.onChange != nil {
		sm.onChange(s, oldState)
	}
	return nil
}

// Fail transitions a session to failed state.
func (sm *SessionManager) Fail(id string, reason string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s, ok := sm.sessions[id]
	if !ok {
		return fmt.Errorf("session %s not found", id)
	}

	oldState := s.State
	s.State = StateFailed
	s.Metadata["failReason"] = reason
	s.StoppedAt = time.Now().UnixMilli()
	if sm.onChange != nil {
		sm.onChange(s, oldState)
	}
	return nil
}

// Get returns a session by ID.
func (sm *SessionManager) Get(id string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	s, ok := sm.sessions[id]
	if !ok {
		return nil, false
	}
	// Return a copy
	copy := *s
	return &copy, true
}

// List returns all sessions.
func (sm *SessionManager) List() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]*Session, 0, len(sm.sessions))
	for _, s := range sm.sessions {
		copy := *s
		result = append(result, &copy)
	}
	return result
}

// ListByState returns sessions filtered by state.
func (sm *SessionManager) ListByState(state SessionState) []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var result []*Session
	for _, s := range sm.sessions {
		if s.State == state {
			copy := *s
			result = append(result, &copy)
		}
	}
	return result
}

// ListActive returns all running/starting sessions.
func (sm *SessionManager) ListActive() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var result []*Session
	for _, s := range sm.sessions {
		if s.State == StateRunning || s.State == StateStarting {
			copy := *s
			result = append(result, &copy)
		}
	}
	return result
}

// UpdateMetadata updates session metadata.
func (sm *SessionManager) UpdateMetadata(id string, metadata map[string]string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s, ok := sm.sessions[id]
	if !ok {
		return fmt.Errorf("session %s not found", id)
	}
	for k, v := range metadata {
		s.Metadata[k] = v
	}
	return nil
}

// Delete removes a session from tracking.
func (sm *SessionManager) Delete(id string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	_, ok := sm.sessions[id]
	if ok {
		delete(sm.sessions, id)
	}
	return ok
}

// Count returns the total number of tracked sessions.
func (sm *SessionManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// Summary returns aggregate session statistics.
func (sm *SessionManager) Summary() *SessionSummary {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	summary := &SessionSummary{
		Total:     len(sm.sessions),
		ByState:   make(map[string]int),
		ByCLIType: make(map[string]int),
	}

	for _, s := range sm.sessions {
		summary.ByState[string(s.State)]++
		summary.ByCLIType[s.CLIType]++
		if s.State == StateRunning {
			summary.ActiveCount++
		}
	}

	return summary
}

// SessionSummary holds aggregate session statistics.
type SessionSummary struct {
	Total       int            `json:"total"`
	ActiveCount int            `json:"activeCount"`
	ByState     map[string]int `json:"byState"`
	ByCLIType   map[string]int `json:"byCLIType"`
}

// --- internal ---

func (sm *SessionManager) transition(id string, newState SessionState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s, ok := sm.sessions[id]
	if !ok {
		return fmt.Errorf("session %s not found", id)
	}

	oldState := s.State
	s.State = newState
	if newState == StateStopped || newState == StateFailed {
		s.StoppedAt = time.Now().UnixMilli()
	}

	if sm.onChange != nil {
		sm.onChange(s, oldState)
	}
	return nil
}
