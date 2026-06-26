package mcp

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// SubmoduleManager manages MCP submodules (git-based MCP server packages).
type SubmoduleManager struct {
	mu            sync.RWMutex
	workspaceRoot string
	submodules    map[string]*Submodule
}

// Submodule represents a registered MCP submodule.
type Submodule struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	URL      string `json:"url"`
	Branch   string `json:"branch"`
	Enabled  bool   `json:"enabled"`
	ToolName string `json:"toolName"`
}

// NewSubmoduleManager creates a new submodule manager.
func NewSubmoduleManager(workspaceRoot string) *SubmoduleManager {
	return &SubmoduleManager{
		workspaceRoot: workspaceRoot,
		submodules:    make(map[string]*Submodule),
	}
}

// Register adds a new submodule.
func (sm *SubmoduleManager) Register(name, url, branch string) error {
	if name == "" || url == "" {
		return fmt.Errorf("name and url are required")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.submodules[name]; ok {
		return fmt.Errorf("submodule already registered: %s", name)
	}

	sm.submodules[name] = &Submodule{
		Name:    name,
		URL:     url,
		Branch:  branch,
		Enabled: true,
	}
	return nil
}

// Clone clones a registered submodule into the workspace.
func (sm *SubmoduleManager) Clone(name string) error {
	sm.mu.RLock()
	sub, ok := sm.submodules[name]
	sm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("submodule not found: %s", name)
	}

	cmd := exec.Command("git", "clone", sub.URL, sub.Path)
	cmd.Dir = sm.workspaceRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone %s: %s: %w", sub.URL, string(output), err)
	}

	if sub.Branch != "" && sub.Branch != "main" && sub.Branch != "master" {
		checkout := exec.Command("git", "checkout", sub.Branch)
		checkout.Dir = sub.Path
		checkout.Run()
	}

	return nil
}

// Update pulls the latest changes for a submodule.
func (sm *SubmoduleManager) Update(name string) error {
	sm.mu.RLock()
	sub, ok := sm.submodules[name]
	sm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("submodule not found: %s", name)
	}

	cmd := exec.Command("git", "pull")
	cmd.Dir = sub.Path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update %s: %s: %w", name, string(output), err)
	}

	return nil
}

// List returns all registered submodules.
func (sm *SubmoduleManager) List() []*Submodule {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]*Submodule, 0, len(sm.submodules))
	for _, sub := range sm.submodules {
		result = append(result, sub)
	}
	return result
}

// Enable enables a submodule.
func (sm *SubmoduleManager) Enable(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sub, ok := sm.submodules[name]
	if !ok {
		return fmt.Errorf("submodule not found: %s", name)
	}
	sub.Enabled = true
	return nil
}

// Disable disables a submodule.
func (sm *SubmoduleManager) Disable(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sub, ok := sm.submodules[name]
	if !ok {
		return fmt.Errorf("submodule not found: %s", name)
	}
	sub.Enabled = false
	return nil
}

// DetectFromPath scans a path for MCP submodule configurations.
func (sm *SubmoduleManager) DetectFromPath(basePath string) ([]string, error) {
	// Simplified detection: look for go.mod or package.json in subdirectories
	// In a full implementation, this would scan recursively
	return nil, nil
}

// GetToolCommand returns the command to run for a submodule's MCP server.
func (sm *SubmoduleManager) GetToolCommand(name string) (string, []string, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sub, ok := sm.submodules[name]
	if !ok {
		return "", nil, fmt.Errorf("submodule not found: %s", name)
	}

	// Default: run the tool via npx or go run
	if strings.Contains(sub.URL, "github.com") {
		repoParts := strings.Split(strings.TrimPrefix(sub.URL, "https://github.com/"), "/")
		if len(repoParts) >= 2 {
			return "npx", []string{"-y", repoParts[1]}, nil
		}
	}

	return "", nil, fmt.Errorf("unknown tool command for submodule: %s", name)
}
