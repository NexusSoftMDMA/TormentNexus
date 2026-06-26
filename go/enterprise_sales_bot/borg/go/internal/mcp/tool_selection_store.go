package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// ToolSelectionStore persists tool selection telemetry to a JSON file.
type ToolSelectionStore struct {
	mu       sync.RWMutex
	filePath string
	maxSize  int
	events   []ToolSelectionEvent
}

// NewToolSelectionStore creates a new file-backed telemetry store.
func NewToolSelectionStore(configDir string, maxSize int) *ToolSelectionStore {
	if maxSize <= 0 {
		maxSize = 1000
	}
	dir := filepath.Join(configDir, "mcp")
	os.MkdirAll(dir, 0755)
	store := &ToolSelectionStore{
		filePath: filepath.Join(dir, "tool_selection_telemetry.json"),
		maxSize:  maxSize,
		events:   make([]ToolSelectionEvent, 0, maxSize),
	}
	store.load()
	return store
}

// GetAll returns all recorded events.
func (s *ToolSelectionStore) GetAll() []ToolSelectionEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]ToolSelectionEvent, len(s.events))
	copy(result, s.events)
	return result
}

// GetStats returns summary statistics.
func (s *ToolSelectionStore) GetStats() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalSelected := 0
	toolCounts := make(map[string]int)
	serverCounts := make(map[string]int)
	for _, e := range s.events {
		if e.Selected {
			totalSelected++
		}
		toolCounts[e.ToolName]++
		serverCounts[e.ServerName]++
	}
	return map[string]any{
		"totalEvents":   len(s.events),
		"totalSelected": totalSelected,
		"uniqueTools":   len(toolCounts),
		"uniqueServers": len(serverCounts),
		"topTools":      toolCounts,
		"topServers":    serverCounts,
	}
}

// Clear removes all events from memory and file.
func (s *ToolSelectionStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = make([]ToolSelectionEvent, 0, s.maxSize)
	return s.save()
}

func (s *ToolSelectionStore) load() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return // File doesn't exist yet — start empty
	}
	var events []ToolSelectionEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return
	}
	if len(events) > s.maxSize {
		events = events[len(events)-s.maxSize:]
	}
	s.events = events
}

func (s *ToolSelectionStore) save() error {
	data, err := json.MarshalIndent(s.events, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}
