package mcp

import (
	"sync"
	"time"
)

// ToolSelectionEvent records a single tool selection event.
type ToolSelectionEvent struct {
	Timestamp      int64   `json:"timestamp"`
	ToolName       string  `json:"toolName"`
	ServerName     string  `json:"serverName"`
	SessionID      string  `json:"sessionId"`
	Score          float64 `json:"score"`
	Selected       bool    `json:"selected"`
	ResponseTimeMs int64   `json:"responseTimeMs,omitempty"`
}

// ToolSelectionTelemetry tracks tool selection events for analysis.
type ToolSelectionTelemetry struct {
	mu      sync.RWMutex
	events  []ToolSelectionEvent
	maxSize int
}

// NewToolSelectionTelemetry creates a new telemetry tracker.
func NewToolSelectionTelemetry(maxSize int) *ToolSelectionTelemetry {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &ToolSelectionTelemetry{
		events:  make([]ToolSelectionEvent, 0, maxSize),
		maxSize: maxSize,
	}
}

// Record adds a tool selection event.
func (tst *ToolSelectionTelemetry) Record(event ToolSelectionEvent) {
	tst.mu.Lock()
	defer tst.mu.Unlock()

	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	tst.events = append(tst.events, event)
	if len(tst.events) > tst.maxSize {
		tst.events = tst.events[len(tst.events)-tst.maxSize:]
	}
}

// GetEvents returns all recorded events.
func (tst *ToolSelectionTelemetry) GetEvents() []ToolSelectionEvent {
	tst.mu.RLock()
	defer tst.mu.RUnlock()

	result := make([]ToolSelectionEvent, len(tst.events))
	copy(result, tst.events)
	return result
}

// GetStats returns summary statistics for tool selection.
func (tst *ToolSelectionTelemetry) GetStats() map[string]interface{} {
	tst.mu.RLock()
	defer tst.mu.RUnlock()

	totalSelected := 0
	toolCounts := make(map[string]int)
	serverCounts := make(map[string]int)

	for _, e := range tst.events {
		if e.Selected {
			totalSelected++
		}
		toolCounts[e.ToolName]++
		serverCounts[e.ServerName]++
	}

	return map[string]interface{}{
		"totalEvents":   len(tst.events),
		"totalSelected": totalSelected,
		"uniqueTools":   len(toolCounts),
		"uniqueServers": len(serverCounts),
		"topTools":      toolCounts,
		"topServers":    serverCounts,
	}
}

// Clear removes all recorded events.
func (tst *ToolSelectionTelemetry) Clear() {
	tst.mu.Lock()
	defer tst.mu.Unlock()
	tst.events = make([]ToolSelectionEvent, 0, tst.maxSize)
}

// Count returns the number of recorded events.
func (tst *ToolSelectionTelemetry) Count() int {
	tst.mu.RLock()
	defer tst.mu.RUnlock()
	return len(tst.events)
}
