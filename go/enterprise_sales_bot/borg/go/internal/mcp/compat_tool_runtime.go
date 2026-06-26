package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// CompatibilityToolRuntime provides a compatibility layer for running tools
// that are defined in the TS compatibility layer.
type CompatibilityToolRuntime struct {
	mu       sync.RWMutex
	handlers map[string]CompatibilityToolHandler
}

// CompatibilityToolHandler is a function that handles a compatibility tool call.
type CompatibilityToolHandler func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// NewCompatibilityToolRuntime creates a new compatibility tool runtime.
func NewCompatibilityToolRuntime() *CompatibilityToolRuntime {
	return &CompatibilityToolRuntime{
		handlers: make(map[string]CompatibilityToolHandler),
	}
}

// RegisterHandler registers a handler for a compatibility tool.
func (rt *CompatibilityToolRuntime) RegisterHandler(name string, handler CompatibilityToolHandler) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.handlers[name] = handler
}

// Execute runs a compatibility tool by name with the given arguments.
func (rt *CompatibilityToolRuntime) Execute(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	rt.mu.RLock()
	handler, ok := rt.handlers[name]
	rt.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("compatibility tool not found: %s", name)
	}

	result, err := handler(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("compatibility tool %s failed: %w", name, err)
	}

	// Serialize to JSON and back to ensure clean shape
	bytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	var clean interface{}
	json.Unmarshal(bytes, &clean)
	return clean, nil
}

// HasHandler checks if a handler is registered for the given tool name.
func (rt *CompatibilityToolRuntime) HasHandler(name string) bool {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	_, ok := rt.handlers[name]
	return ok
}

// ListHandlers returns the names of all registered handlers.
func (rt *CompatibilityToolRuntime) ListHandlers() []string {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	names := make([]string, 0, len(rt.handlers))
	for n := range rt.handlers {
		names = append(names, n)
	}
	return names
}
