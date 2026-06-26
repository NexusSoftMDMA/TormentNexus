// Package workspaces provides workspace registration and discovery
// ported from packages/core/src/services/WorkspaceTracker.ts.
//
// It maintains a registry of recently-accessed workspaces in
// ~/.tormentnexus/workspaces.json, verifying they still exist on disk.
package workspaces

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"
)

// WorkspaceRecord represents a single registered workspace.
type WorkspaceRecord struct {
	Path          string `json:"path"`
	LastAccessedAt int64  `json:"lastAccessedAt"`
	Name          string `json:"name"`
}

// WorkspaceTracker manages the workspace registry.
type WorkspaceTracker struct {
	mu           sync.Mutex
	registryPath string
	maxRecords   int
}

// NewWorkspaceTracker creates a new tracker.
// If registryPath is empty, defaults to ~/.tormentnexus/workspaces.json.
func NewWorkspaceTracker(registryPath string) *WorkspaceTracker {
	if registryPath == "" {
		home, _ := os.UserHomeDir()
		if home == "" {
			home = "."
		}
		registryPath = filepath.Join(home, ".tormentnexus", "workspaces.json")
	}
	return &WorkspaceTracker{
		registryPath: registryPath,
		maxRecords:   50,
	}
}

// RegisterWorkspace adds or updates a workspace in the registry.
func (wt *WorkspaceTracker) RegisterWorkspace(workspacePath string) error {
	if workspacePath == "" {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		workspacePath = dir
	}

	wt.mu.Lock()
	defer wt.mu.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(wt.registryPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing
	workspaces := wt.load()

	// Remove existing entry
	filtered := make([]WorkspaceRecord, 0, len(workspaces))
	for _, ws := range workspaces {
		if ws.Path != workspacePath {
			filtered = append(filtered, ws)
		}
	}

	// Prepend new entry
	record := WorkspaceRecord{
		Path:          workspacePath,
		Name:          filepath.Base(workspacePath),
		LastAccessedAt: time.Now().UnixMilli(),
	}
	workspaces = append([]WorkspaceRecord{record}, filtered...)

	// Keep top maxRecords
	if len(workspaces) > wt.maxRecords {
		workspaces = workspaces[:wt.maxRecords]
	}

	// Write back
	return wt.save(workspaces)
}

// ListWorkspaces returns all registered workspaces that still exist on disk.
func (wt *WorkspaceTracker) ListWorkspaces() ([]WorkspaceRecord, error) {
	wt.mu.Lock()
	defer wt.mu.Unlock()

	workspaces := wt.load()

	// Verify they still exist on disk
	var valid []WorkspaceRecord
	for _, ws := range workspaces {
		info, err := os.Stat(ws.Path)
		if err == nil && info.IsDir() {
			valid = append(valid, ws)
		}
	}

	return valid, nil
}

// GetRecent returns the N most recently accessed workspaces.
func (wt *WorkspaceTracker) GetRecent(n int) ([]WorkspaceRecord, error) {
	all, err := wt.ListWorkspaces()
	if err != nil {
		return nil, err
	}
	if n > 0 && n < len(all) {
		return all[:n], nil
	}
	return all, nil
}

// Remove deletes a workspace from the registry.
func (wt *WorkspaceTracker) Remove(path string) error {
	wt.mu.Lock()
	defer wt.mu.Unlock()

	workspaces := wt.load()
	filtered := make([]WorkspaceRecord, 0, len(workspaces))
	for _, ws := range workspaces {
		if ws.Path != path {
			filtered = append(filtered, ws)
		}
	}
	return wt.save(filtered)
}

// --- internal ---

func (wt *WorkspaceTracker) load() []WorkspaceRecord {
	data, err := os.ReadFile(wt.registryPath)
	if err != nil {
		return nil
	}
	var workspaces []WorkspaceRecord
	if err := json.Unmarshal(data, &workspaces); err != nil {
		return nil
	}
	return workspaces
}

func (wt *WorkspaceTracker) save(workspaces []WorkspaceRecord) error {
	data, err := json.MarshalIndent(workspaces, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workspaces: %w", err)
	}
	return os.WriteFile(wt.registryPath, data, 0644)
}

// Ensure unused imports are consumed
var (
	_ = sort.Slice
	_ = runtime.NumCPU
)
