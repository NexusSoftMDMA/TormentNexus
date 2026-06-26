// Package toolregistry provides tool registration and lookup ported from
// packages/core/src/services/ToolRegistry.ts.
//
// It tracks all known tools, their schemas, and provides fast lookup by name.
package toolregistry

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ToolInfo describes a registered tool.
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ServerName  string                 `json:"serverName,omitempty"`
	Category    string                 `json:"category,omitempty"`
	AlwaysOn    bool                   `json:"alwaysOn"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Source      string                 `json:"source,omitempty"` // "native", "bridge", "discovered"
}

// ToolRegistry manages the global tool inventory.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]*ToolInfo
}

// NewToolRegistry creates a new empty registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]*ToolInfo),
	}
}

// Register adds or updates a tool in the registry.
func (tr *ToolRegistry) Register(tool ToolInfo) error {
	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.tools[strings.ToLower(tool.Name)] = &tool
	return nil
}

// RegisterBatch registers multiple tools at once.
func (tr *ToolRegistry) RegisterBatch(tools []ToolInfo) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	for _, tool := range tools {
		if tool.Name == "" {
			return fmt.Errorf("tool name cannot be empty")
		}
		tr.tools[strings.ToLower(tool.Name)] = &tool
	}
	return nil
}

// Unregister removes a tool from the registry.
func (tr *ToolRegistry) Unregister(name string) bool {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	key := strings.ToLower(name)
	_, existed := tr.tools[key]
	if existed {
		delete(tr.tools, key)
	}
	return existed
}

// Get returns a tool by name (case-insensitive).
func (tr *ToolRegistry) Get(name string) (*ToolInfo, bool) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	t, ok := tr.tools[strings.ToLower(name)]
	if !ok {
		return nil, false
	}
	copy := *t
	return &copy, true
}

// List returns all registered tools.
func (tr *ToolRegistry) List() []ToolInfo {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	result := make([]ToolInfo, 0, len(tr.tools))
	for _, t := range tr.tools {
		result = append(result, *t)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ListByServer returns tools filtered by server name.
func (tr *ToolRegistry) ListByServer(serverName string) []ToolInfo {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	var result []ToolInfo
	for _, t := range tr.tools {
		if t.ServerName == serverName {
			result = append(result, *t)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ListAlwaysOn returns tools marked as always-on.
func (tr *ToolRegistry) ListAlwaysOn() []ToolInfo {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	var result []ToolInfo
	for _, t := range tr.tools {
		if t.AlwaysOn {
			result = append(result, *t)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// Search performs a fuzzy search across tool names, descriptions, and tags.
func (tr *ToolRegistry) Search(query string, limit int) []ToolInfo {
	if limit <= 0 {
		limit = 20
	}
	query = strings.ToLower(query)

	tr.mu.RLock()
	defer tr.mu.RUnlock()

	type scored struct {
		tool  ToolInfo
		score float64
	}

	var results []scored
	for _, t := range tr.tools {
		score := 0.0

		// Exact name match
		if strings.ToLower(t.Name) == query {
			score += 10.0
		}

		// Name prefix
		if strings.HasPrefix(strings.ToLower(t.Name), query) {
			score += 5.0
		}

		// Name contains
		if strings.Contains(strings.ToLower(t.Name), query) {
			score += 3.0
		}

		// Description match
		if strings.Contains(strings.ToLower(t.Description), query) {
			score += 2.0
		}

		// Tag match
		for _, tag := range t.Tags {
			if strings.ToLower(tag) == query {
				score += 4.0
			}
			if strings.Contains(strings.ToLower(tag), query) {
				score += 1.0
			}
		}

		// Category match
		if strings.Contains(strings.ToLower(t.Category), query) {
			score += 1.5
		}

		if score > 0 {
			results = append(results, scored{tool: *t, score: score})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	tools := make([]ToolInfo, len(results))
	for i, r := range results {
		tools[i] = r.tool
	}
	return tools
}

// SetAlwaysOn enables or disables the always-on flag for a tool.
func (tr *ToolRegistry) SetAlwaysOn(name string, alwaysOn bool) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	key := strings.ToLower(name)
	t, ok := tr.tools[key]
	if !ok {
		return fmt.Errorf("tool %s not found", name)
	}
	t.AlwaysOn = alwaysOn
	return nil
}

// Count returns the total number of registered tools.
func (tr *ToolRegistry) Count() int {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	return len(tr.tools)
}

// Stats returns aggregate statistics.
func (tr *ToolRegistry) Stats() *RegistryStats {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	stats := &RegistryStats{
		Total:      len(tr.tools),
		ByCategory: make(map[string]int),
		ByServer:   make(map[string]int),
		BySource:   make(map[string]int),
	}

	for _, t := range tr.tools {
		stats.ByCategory[t.Category]++
		stats.ByServer[t.ServerName]++
		stats.BySource[t.Source]++
		if t.AlwaysOn {
			stats.AlwaysOnCount++
		}
	}

	return stats
}

// RegistryStats holds aggregate tool registry statistics.
type RegistryStats struct {
	Total         int            `json:"total"`
	AlwaysOnCount int            `json:"alwaysOnCount"`
	ByCategory    map[string]int `json:"byCategory"`
	ByServer      map[string]int `json:"byServer"`
	BySource      map[string]int `json:"bySource"`
}

// Clear removes all tools from the registry.
func (tr *ToolRegistry) Clear() {
	tr.mu.Lock()
	tr.tools = make(map[string]*ToolInfo)
	tr.mu.Unlock()
}
