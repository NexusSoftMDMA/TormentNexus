// Package eventbus provides a typed, thread-safe, pattern-matching event bus
// ported from packages/core/src/services/EventBus.ts.
//
// It supports exact-match subscriptions, wildcard patterns (e.g. "file:*"),
// bounded history, and safe concurrent access.
package eventbus

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

// SystemEventType enumerates well-known event categories.
type SystemEventType string

const (
	EventAgentHeartbeat SystemEventType = "agent:heartbeat"
	EventAgentStart     SystemEventType = "agent:start"
	EventAgentStop      SystemEventType = "agent:stop"
	EventTaskUpdate     SystemEventType = "task:update"
	EventTaskComplete   SystemEventType = "task:complete"
	EventToolCall       SystemEventType = "tool:call"
	EventMemoryPrune    SystemEventType = "memory:prune"
	EventFileChange     SystemEventType = "file:change"
	EventTerminalError  SystemEventType = "terminal:error"
	EventA2ASignal      SystemEventType = "a2a:signal"
	EventUserActivity   SystemEventType = "user:activity"
)

// SystemEvent is a single structured event emitted by any component.
type SystemEvent struct {
	Type      SystemEventType `json:"type"`
	Timestamp int64           `json:"timestamp"`
	Source    string          `json:"source"`
	Payload   interface{}     `json:"payload,omitempty"`
}

// wildcardListener binds a compiled glob pattern to a callback.
type wildcardListener struct {
	pattern  *regexp.Regexp
	listener func(SystemEvent)
}

// EventBus is a typed, thread-safe, pattern-matching event bus.
type EventBus struct {
	mu                 sync.RWMutex
	exactListeners     map[string][]func(SystemEvent)
	wildcardListeners  []wildcardListener
	globalListeners    []func(SystemEvent)
	history            []SystemEvent
	maxHistory         int
}

// New creates a new EventBus with the given maximum history size.
func New(maxHistory int) *EventBus {
	if maxHistory <= 0 {
		maxHistory = 1000
	}
	return &EventBus{
		exactListeners: make(map[string][]func(SystemEvent)),
		history:        make([]SystemEvent, 0, maxHistory),
		maxHistory:     maxHistory,
	}
}

// Subscribe registers a listener for events matching pattern.
// If pattern contains '*', it is treated as a wildcard (e.g. "file:*" matches "file:change").
func (eb *EventBus) Subscribe(pattern string, listener func(SystemEvent)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if containsGlob(pattern) {
		regex := globToRegex(pattern)
		eb.wildcardListeners = append(eb.wildcardListeners, wildcardListener{
			pattern:  regex,
			listener: listener,
		})
	} else {
		eb.exactListeners[pattern] = append(eb.exactListeners[pattern], listener)
	}
}

// OnGlobal registers a listener that fires for every event.
func (eb *EventBus) OnGlobal(listener func(SystemEvent)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.globalListeners = append(eb.globalListeners, listener)
}

// EmitEvent publishes an event to all matching subscribers.
func (eb *EventBus) EmitEvent(eventType SystemEventType, source string, payload interface{}) {
	event := SystemEvent{
		Type:      eventType,
		Timestamp: time.Now().UnixMilli(),
		Source:    source,
		Payload:   payload,
	}

	eb.mu.Lock()
	// Store in bounded history
	eb.history = append(eb.history, event)
	if len(eb.history) > eb.maxHistory {
		eb.history = eb.history[len(eb.history)-eb.maxHistory:]
	}
	// Snapshot listeners under write lock
	exact := eb.exactListeners[string(eventType)]
	var wcs []wildcardListener
	wcs = append(wcs, eb.wildcardListeners...)
	var globals []func(SystemEvent)
	globals = append(globals, eb.globalListeners...)
	eb.mu.Unlock()

	// Fire global listeners
	for _, fn := range globals {
		safeCall(fn, event)
	}

	// Fire exact-match listeners
	for _, fn := range exact {
		safeCall(fn, event)
	}

	// Fire wildcard listeners
	for _, wl := range wcs {
		if wl.pattern.MatchString(string(eventType)) {
			safeCall(wl.listener, event)
		}
	}
}

// GetHistory returns the last `limit` events from history.
func (eb *EventBus) GetHistory(limit int) []SystemEvent {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if limit <= 0 || limit > len(eb.history) {
		limit = len(eb.history)
	}

	start := len(eb.history) - limit
	if start < 0 {
		start = 0
	}

	result := make([]SystemEvent, len(eb.history)-start)
	copy(result, eb.history[start:])
	return result
}

// GetHistorySince returns all events in history that occurred after the given timestamp.
func (eb *EventBus) GetHistorySince(timestamp int64) []SystemEvent {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	var result []SystemEvent
	for _, ev := range eb.history {
		if ev.Timestamp > timestamp {
			result = append(result, ev)
		}
	}
	return result
}

// --- helpers ---

func containsGlob(s string) bool {
	for _, ch := range s {
		if ch == '*' {
			return true
		}
	}
	return false
}

// globToRegex converts a simple glob pattern ("file:*") to a regex.
func globToRegex(glob string) *regexp.Regexp {
	escaped := regexp.QuoteMeta(glob)
	// QuoteMeta escapes '*' to '\*', so replace that back with '.*'
	regex := "^" + strings.ReplaceAll(escaped, `\*`, ".*") + "$"
	re, _ := regexp.Compile(regex)
	if re == nil {
		// Fallback: match everything
		re = regexp.MustCompile(".*")
	}
	return re
}

func safeCall(fn func(SystemEvent), event SystemEvent) {
	defer func() {
		recover() //nolint:errcheck
	}()
	fn(event)
}
