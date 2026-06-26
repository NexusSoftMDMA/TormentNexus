package mcp

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// MCPTrafficDirection indicates the direction of MCP traffic.
type MCPTrafficDirection string

const (
	TrafficDirectionIncoming MCPTrafficDirection = "incoming"
	TrafficDirectionOutgoing MCPTrafficDirection = "outgoing"
)

// MCPTrafficEvent represents a single MCP traffic event (request or response).
type MCPTrafficEvent struct {
	ID         string              `json:"id"`
	Timestamp  string              `json:"timestamp"`
	Direction  MCPTrafficDirection `json:"direction"`
	ServerName string              `json:"serverName"`
	Method     string              `json:"method"`
	Params     string              `json:"params,omitempty"`
	Result     string              `json:"result,omitempty"`
	Error      string              `json:"error,omitempty"`
	DurationMs int64               `json:"durationMs"`
}

// MCPTrafficInspector captures and stores MCP traffic events in a ring buffer.
type MCPTrafficInspector struct {
	mu        sync.RWMutex
	events    []MCPTrafficEvent
	maxEvents int
	nextID    int
}

// NewMCPTrafficInspector creates a new traffic inspector with the given max capacity.
func NewMCPTrafficInspector(maxEvents int) *MCPTrafficInspector {
	if maxEvents <= 0 {
		maxEvents = 200
	}
	return &MCPTrafficInspector{
		events:    make([]MCPTrafficEvent, 0, maxEvents),
		maxEvents: maxEvents,
	}
}

// Record appends a traffic event to the ring buffer.
func (ti *MCPTrafficInspector) Record(event MCPTrafficEvent) {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	ti.nextID++
	if event.ID == "" {
		event.ID = fmt.Sprintf("evt-%d", ti.nextID)
	}
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}

	ti.events = append(ti.events, event)
	if len(ti.events) > ti.maxEvents {
		ti.events = ti.events[len(ti.events)-ti.maxEvents:]
	}
}

// RecordRequest records an outgoing MCP request.
func (ti *MCPTrafficInspector) RecordRequest(serverName, method string, params interface{}) MCPTrafficEvent {
	event := MCPTrafficEvent{
		Direction:  TrafficDirectionOutgoing,
		ServerName: serverName,
		Method:     method,
		Params:     formatParams(params),
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
	}
	ti.Record(event)
	return event
}

// RecordResponse records an incoming MCP response, linked to a prior request.
func (ti *MCPTrafficInspector) RecordResponse(serverName, method string, result interface{}, err error, durationMs int64) {
	event := MCPTrafficEvent{
		Direction:  TrafficDirectionIncoming,
		ServerName: serverName,
		Method:     method,
		DurationMs: durationMs,
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
	}
	if err != nil {
		event.Error = err.Error()
	} else {
		event.Result = formatParams(result)
	}
	ti.Record(event)
}

// GetEvents returns a copy of all recorded events.
func (ti *MCPTrafficInspector) GetEvents() []MCPTrafficEvent {
	ti.mu.RLock()
	defer ti.mu.RUnlock()

	result := make([]MCPTrafficEvent, len(ti.events))
	copy(result, ti.events)
	return result
}

// GetEventsByServer returns events filtered by server name.
func (ti *MCPTrafficInspector) GetEventsByServer(serverName string) []MCPTrafficEvent {
	ti.mu.RLock()
	defer ti.mu.RUnlock()

	var result []MCPTrafficEvent
	for _, e := range ti.events {
		if e.ServerName == serverName {
			result = append(result, e)
		}
	}
	return result
}

// GetEventsByMethod returns events filtered by MCP method.
func (ti *MCPTrafficInspector) GetEventsByMethod(method string) []MCPTrafficEvent {
	ti.mu.RLock()
	defer ti.mu.RUnlock()

	var result []MCPTrafficEvent
	for _, e := range ti.events {
		if e.Method == method {
			result = append(result, e)
		}
	}
	return result
}

// Clear removes all recorded events.
func (ti *MCPTrafficInspector) Clear() {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	ti.events = make([]MCPTrafficEvent, 0, ti.maxEvents)
}

// EventCount returns the number of recorded events.
func (ti *MCPTrafficInspector) EventCount() int {
	ti.mu.RLock()
	defer ti.mu.RUnlock()
	return len(ti.events)
}

// formatParams converts parameters to a human-readable summary string.
func formatParams(v interface{}) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		if len(val) > 80 {
			return val[:77] + "..."
		}
		return val
	case []byte:
		s := string(val)
		if len(s) > 80 {
			return s[:77] + "..."
		}
		return s
	case fmt.Stringer:
		return val.String()
	default:
		s := fmt.Sprintf("%v", v)
		if len(s) > 80 {
			return s[:77] + "..."
		}
		return s
	}
}

// SummarizeParams creates a compact parameter summary like "key1=val1, key2=val2".
func SummarizeParams(args map[string]interface{}) string {
	if len(args) == 0 {
		return ""
	}

	var parts []string
	count := 0
	for k, v := range args {
		if count >= 5 {
			break
		}
		parts = append(parts, fmt.Sprintf("%s=%s", k, formatPrimitive(v)))
		count++
	}
	return strings.Join(parts, ", ")
}

func formatPrimitive(v interface{}) string {
	switch val := v.(type) {
	case string:
		if len(val) > 40 {
			return val[:37] + "..."
		}
		return val
	case float64, float32, int, int64, bool:
		return fmt.Sprintf("%v", v)
	case nil:
		return "null"
	case []interface{}:
		return fmt.Sprintf("[%d items]", len(val))
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
			if len(keys) >= 4 {
				break
			}
		}
		return "{" + strings.Join(keys, ", ") + "}"
	default:
		return fmt.Sprintf("%v", v)
	}
}
