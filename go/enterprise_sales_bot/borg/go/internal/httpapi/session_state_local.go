package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type localSessionState struct {
	IsAutoDriveActive bool    `json:"isAutoDriveActive"`
	ActiveGoal        *string `json:"activeGoal"`
	LastObjective     *string `json:"lastObjective"`
	LastHeartbeat     int64   `json:"lastHeartbeat"`
	ThreadID          *string `json:"threadId,omitempty"`
}

type localSessionStateManager struct {
	path  string
	mu    sync.RWMutex
	state localSessionState
}

func newLocalSessionStateManager(path string) *localSessionStateManager {
	manager := &localSessionStateManager{
		path:  path,
		state: defaultLocalSessionState(),
	}
	manager.load()
	return manager
}

func defaultLocalSessionState() localSessionState {
	return localSessionState{
		IsAutoDriveActive: false,
		ActiveGoal:        nil,
		LastObjective:     nil,
		LastHeartbeat:     time.Now().UnixMilli(),
		ThreadID:          nil,
	}
}

func (m *localSessionStateManager) load() {
	m.mu.Lock()
	defer m.mu.Unlock()

	raw, err := os.ReadFile(m.path)
	if err != nil {
		return
	}

	var state localSessionState
	if err := json.Unmarshal(raw, &state); err != nil {
		return
	}
	if state.LastHeartbeat == 0 {
		state.LastHeartbeat = time.Now().UnixMilli()
	}
	m.state = state
}

func (m *localSessionStateManager) snapshot() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneLocalSessionState(m.state)
}

func (m *localSessionStateManager) update(patch map[string]any) map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()

	if raw, ok := patch["isAutoDriveActive"]; ok {
		if value, ok := raw.(bool); ok {
			m.state.IsAutoDriveActive = value
		}
	}
	if raw, ok := patch["activeGoal"]; ok {
		m.state.ActiveGoal = nullableTrimmedString(raw)
	}
	if raw, ok := patch["lastObjective"]; ok {
		m.state.LastObjective = nullableTrimmedString(raw)
	}
	if raw, ok := patch["threadId"]; ok {
		m.state.ThreadID = nullableTrimmedString(raw)
	}
	m.state.LastHeartbeat = time.Now().UnixMilli()
	m.saveLocked()
	return cloneLocalSessionState(m.state)
}

func (m *localSessionStateManager) clear() map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = defaultLocalSessionState()
	m.saveLocked()
	return cloneLocalSessionState(m.state)
}

func (m *localSessionStateManager) touch() map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.LastHeartbeat = time.Now().UnixMilli()
	m.saveLocked()
	return cloneLocalSessionState(m.state)
}

func (m *localSessionStateManager) saveLocked() {
	if strings.TrimSpace(m.path) == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(m.path), 0o755); err != nil {
		return
	}
	raw, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(m.path, raw, 0o644)
}

func cloneLocalSessionState(state localSessionState) map[string]any {
	result := map[string]any{
		"isAutoDriveActive": state.IsAutoDriveActive,
		"activeGoal":        nullableStringValue(state.ActiveGoal),
		"lastObjective":     nullableStringValue(state.LastObjective),
		"lastHeartbeat":     state.LastHeartbeat,
	}
	if state.ThreadID != nil && strings.TrimSpace(*state.ThreadID) != "" {
		result["threadId"] = strings.TrimSpace(*state.ThreadID)
	}
	return result
}

func nullableTrimmedString(value any) *string {
	switch typed := value.(type) {
	case nil:
		return nil
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return nil
		}
		return &trimmed
	default:
		return nil
	}
}

func nullableStringValue(value *string) any {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}
