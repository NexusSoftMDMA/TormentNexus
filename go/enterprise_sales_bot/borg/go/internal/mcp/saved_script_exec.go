package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// SavedScriptExecution handles execution of previously saved scripts through MCP.
type SavedScriptExecution struct {
	mu       sync.RWMutex
	scripts  map[string]string // name -> code
	executor ScriptExecutor
}

// ScriptExecutor is a function that executes script code and returns the result.
type ScriptExecutor func(ctx context.Context, code string, args map[string]interface{}) (string, error)

// NewSavedScriptExecution creates a new saved script execution handler.
func NewSavedScriptExecution(executor ScriptExecutor) *SavedScriptExecution {
	if executor == nil {
		executor = defaultScriptExecutor
	}
	return &SavedScriptExecution{
		scripts:  make(map[string]string),
		executor: executor,
	}
}

// Save stores a script by name.
func (sse *SavedScriptExecution) Save(name, code string) error {
	if name == "" {
		return fmt.Errorf("script name is required")
	}
	if code == "" {
		return fmt.Errorf("script code is required")
	}

	sse.mu.Lock()
	defer sse.mu.Unlock()
	sse.scripts[name] = code
	return nil
}

// Get retrieves a saved script by name.
func (sse *SavedScriptExecution) Get(name string) (string, error) {
	sse.mu.RLock()
	defer sse.mu.RUnlock()

	code, ok := sse.scripts[name]
	if !ok {
		return "", fmt.Errorf("script not found: %s", name)
	}
	return code, nil
}

// Execute runs a saved script with the given arguments.
func (sse *SavedScriptExecution) Execute(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	sse.mu.RLock()
	code, ok := sse.scripts[name]
	sse.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("script not found: %s", name)
	}

	return sse.executor(ctx, code, args)
}

// List returns the names of all saved scripts.
func (sse *SavedScriptExecution) List() []string {
	sse.mu.RLock()
	defer sse.mu.RUnlock()

	names := make([]string, 0, len(sse.scripts))
	for n := range sse.scripts {
		names = append(names, n)
	}
	return names
}

// Delete removes a saved script.
func (sse *SavedScriptExecution) Delete(name string) {
	sse.mu.Lock()
	defer sse.mu.Unlock()
	delete(sse.scripts, name)
}

// defaultScriptExecutor is a placeholder executor that returns the code as-is.
func defaultScriptExecutor(ctx context.Context, code string, args map[string]interface{}) (string, error) {
	argsJSON, _ := json.Marshal(args)
	return fmt.Sprintf("Script executed (placeholder): %s\nArgs: %s", code[:min(len(code), 100)], string(argsJSON)), nil
}
