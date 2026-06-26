package mcp

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// CachedMcpServerInventory holds enriched server metadata for the cached inventory.
type CachedMcpServerInventory struct {
	UUID               string            `json:"uuid"`
	Name               string            `json:"name"`
	Type               string            `json:"type"`
	Command            string            `json:"command"`
	Args               []string          `json:"args"`
	Env                map[string]string `json:"env"`
	URL                string            `json:"url"`
	Description        string            `json:"description"`
	Enabled            bool              `json:"enabled"`
	AlwaysOn           bool              `json:"alwaysOn"`
	DisplayName        string            `json:"displayName"`
	Tags               []string          `json:"tags"`
	AlwaysOnAdvertised bool              `json:"alwaysOnAdvertised"`
	Source             string            `json:"source"`      // "config" or "database"
	ErrorStatus        string            `json:"errorStatus"` // "NONE" or "ERROR"
	ErrorMessage       string            `json:"errorMessage,omitempty"`
	DiscoveredAt       string            `json:"discoveredAt,omitempty"`
	UpdatedAt          string            `json:"updatedAt,omitempty"`
}

// CachedMcpToolInventory holds enriched tool metadata for the cached inventory.
type CachedMcpToolInventory struct {
	Name               string      `json:"name"`
	Description        string      `json:"description"`
	Server             string      `json:"server"`
	ServerDisplayName  string      `json:"serverDisplayName"`
	ServerTags         []string    `json:"serverTags"`
	ToolTags           []string    `json:"toolTags"`
	SemanticGroup      string      `json:"semanticGroup"`
	SemanticGroupLabel string      `json:"semanticGroupLabel"`
	AdvertisedName     string      `json:"advertisedName"`
	Keywords           []string    `json:"keywords"`
	AlwaysOn           bool        `json:"alwaysOn"`
	OriginalName       string      `json:"originalName"`
	InputSchema        interface{} `json:"inputSchema"`
}

// InventorySource indicates where the inventory was sourced from.
type InventorySource string

const (
	InventorySourceDatabase InventorySource = "database"
	InventorySourceConfig   InventorySource = "config"
	InventorySourceEmpty    InventorySource = "empty"
)

// CachedInventorySnapshot is a complete point-in-time view of all MCP servers and tools.
type CachedInventorySnapshot struct {
	Servers           []CachedMcpServerInventory `json:"servers"`
	ToolCounts        map[string]int             `json:"toolCounts"` // keyed by server name
	Tools             []CachedMcpToolInventory   `json:"tools"`
	Source            InventorySource            `json:"source"`
	SnapshotUpdatedAt string                     `json:"snapshotUpdatedAt,omitempty"`
	DatabaseAvailable bool                       `json:"databaseAvailable"`
	DatabaseError     string                     `json:"databaseError,omitempty"`
	FallbackUsed      bool                       `json:"fallbackUsed"`
	CachedAt          string                     `json:"cachedAt"`
}

// CachedInventory manages a cached view of all MCP servers and their tools,
// merging config-based servers with database-persisted servers.
type CachedInventory struct {
	mu       sync.RWMutex
	snapshot *CachedInventorySnapshot
	ttl      time.Duration
	lastLoad time.Time
}

// NewCachedInventory creates a new cached inventory with the given TTL.
func NewCachedInventory(ttl time.Duration) *CachedInventory {
	return &CachedInventory{
		ttl: ttl,
	}
}

// GetSnapshot returns the current inventory snapshot, refreshing it if stale.
func (ci *CachedInventory) GetSnapshot() (*CachedInventorySnapshot, error) {
	ci.mu.RLock()
	if ci.snapshot != nil && time.Since(ci.lastLoad) < ci.ttl {
		snap := ci.snapshot
		ci.mu.RUnlock()
		return snap, nil
	}
	ci.mu.RUnlock()

	return ci.Refresh()
}

// Refresh forces a reload of the inventory from all sources.
func (ci *CachedInventory) Refresh() (*CachedInventorySnapshot, error) {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	configSnapshot := buildConfigSnapshot()
	databaseSnapshot, dbErr := buildDatabaseSnapshot()

	var mergedServers []CachedMcpServerInventory
	var mergedTools []CachedMcpToolInventory
	mergedToolCounts := make(map[string]int)

	// Merge config snapshot
	for _, s := range configSnapshot.Servers {
		mergedServers = append(mergedServers, s)
	}
	for name, count := range configSnapshot.ToolCounts {
		mergedToolCounts[strings.TrimPrefix(name, "config:")] = count
	}
	for _, t := range configSnapshot.Tools {
		mergedTools = append(mergedTools, t)
	}

	if databaseSnapshot != nil {
		for _, s := range databaseSnapshot.Servers {
			mergedServers = append(mergedServers, s)
		}
		for _, t := range databaseSnapshot.Tools {
			mergedTools = append(mergedTools, t)
		}
		for _, s := range databaseSnapshot.Servers {
			if count, ok := databaseSnapshot.ToolCounts[s.UUID]; ok {
				mergedToolCounts[s.Name] = count
			}
		}
	}

	source := InventorySourceEmpty
	if len(mergedTools) > 0 || len(mergedServers) > 0 {
		source = InventorySourceDatabase
	}

	var updatedAt string
	if databaseSnapshot != nil && databaseSnapshot.SnapshotUpdatedAt != "" {
		updatedAt = databaseSnapshot.SnapshotUpdatedAt
	} else if configSnapshot.SnapshotUpdatedAt != "" {
		updatedAt = configSnapshot.SnapshotUpdatedAt
	}

	dbAvailable := true
	var dbError string
	if dbErr != nil {
		dbAvailable = false
		dbError = dbErr.Error()
	}

	ci.snapshot = &CachedInventorySnapshot{
		Servers:           mergedServers,
		ToolCounts:        mergedToolCounts,
		Tools:             mergedTools,
		Source:            source,
		SnapshotUpdatedAt: updatedAt,
		DatabaseAvailable: dbAvailable,
		DatabaseError:     dbError,
		FallbackUsed:      dbErr != nil,
		CachedAt:          time.Now().UTC().Format(time.RFC3339),
	}
	ci.lastLoad = time.Now()

	return ci.snapshot, nil
}

// FindTools searches the cached inventory for tools matching the given criteria.
func (ci *CachedInventory) FindTools(serverName string, toolName string) []CachedMcpToolInventory {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	if ci.snapshot == nil {
		return nil
	}

	var results []CachedMcpToolInventory
	for _, t := range ci.snapshot.Tools {
		if serverName != "" && t.Server != serverName {
			continue
		}
		if toolName != "" && !strings.Contains(strings.ToLower(t.Name), strings.ToLower(toolName)) {
			continue
		}
		results = append(results, t)
	}
	return results
}

// FindServer returns a server by name from the cached inventory.
func (ci *CachedInventory) FindServer(name string) *CachedMcpServerInventory {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	if ci.snapshot == nil {
		return nil
	}

	for _, s := range ci.snapshot.Servers {
		if s.Name == name {
			return &s
		}
	}
	return nil
}

// MarshalJSON serializes the current snapshot to JSON.
func (ci *CachedInventory) MarshalJSON() ([]byte, error) {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	if ci.snapshot == nil {
		return json.Marshal(map[string]string{"status": "empty"})
	}
	return json.Marshal(ci.snapshot)
}

// --- Internal snapshot builders ---

func buildConfigSnapshot() *CachedInventorySnapshot {
	// Load from tormentnexus.config.json / mcp.jsonc
	config, err := LoadMcpJsonConfig()
	if err != nil {
		return &CachedInventorySnapshot{
			Source:   InventorySourceEmpty,
			CachedAt: time.Now().UTC().Format(time.RFC3339),
		}
	}

	var servers []CachedMcpServerInventory
	var tools []CachedMcpToolInventory
	toolCounts := make(map[string]int)
	var updatedAtCandidates []string
	index := 0

	for name, entry := range config.McpServers {
		meta := entry.Meta
		server := CachedMcpServerInventory{
			UUID:               fmt.Sprintf("config:%s:%d", name, index),
			Name:               name,
			Type:               entry.Type,
			Command:            entry.Command,
			Args:               entry.Args,
			Env:                entry.Env,
			URL:                entry.URL,
			Description:        meta.Description,
			Enabled:            !entry.Disabled,
			AlwaysOn:           meta.AlwaysOn,
			DisplayName:        firstNonEmpty(meta.DisplayName, meta.ServerName, name),
			Tags:               meta.ServerTags,
			AlwaysOnAdvertised: meta.AlwaysOn,
			Source:             "config",
			ErrorStatus:        "NONE",
		}

		if meta.Status == "failed" {
			server.ErrorStatus = "ERROR"
			server.ErrorMessage = meta.Error
		}
		if meta.CacheHydratedAt != "" {
			updatedAtCandidates = append(updatedAtCandidates, meta.CacheHydratedAt)
		}
		if meta.DiscoveredAt != "" {
			updatedAtCandidates = append(updatedAtCandidates, meta.DiscoveredAt)
		}

		servers = append(servers, server)
		toolCounts[server.UUID] = len(meta.Tools)

		for _, tool := range meta.Tools {
			nsName := NamespaceToolName(name, tool.Name)
			tools = append(tools, CachedMcpToolInventory{
				Name:               nsName,
				Description:        tool.Description,
				Server:             name,
				ServerDisplayName:  firstNonEmpty(tool.ServerDisplayName, server.DisplayName, name),
				ServerTags:         firstNonEmptySlice(tool.ServerTags, meta.ServerTags),
				ToolTags:           tool.ToolTags,
				SemanticGroup:      firstNonEmpty(tool.SemanticGroup, "general-utility"),
				SemanticGroupLabel: firstNonEmpty(tool.SemanticGroupLabel, "general utility"),
				AdvertisedName:     firstNonEmpty(tool.AdvertisedName, nsName),
				Keywords:           tool.Keywords,
				AlwaysOn:           tool.AlwaysOn || meta.AlwaysOn,
				OriginalName:       tool.Name,
				InputSchema:        tool.InputSchema,
			})
		}
		index++
	}

	sort.Slice(updatedAtCandidates, func(i, j int) bool {
		return updatedAtCandidates[i] > updatedAtCandidates[j]
	})
	var snapshotUpdatedAt string
	if len(updatedAtCandidates) > 0 {
		snapshotUpdatedAt = updatedAtCandidates[0]
	}

	source := InventorySourceEmpty
	if len(tools) > 0 || len(servers) > 0 {
		source = InventorySourceConfig
	}

	return &CachedInventorySnapshot{
		Servers:           servers,
		ToolCounts:        toolCounts,
		Tools:             tools,
		Source:            source,
		SnapshotUpdatedAt: snapshotUpdatedAt,
		DatabaseAvailable: true,
		CachedAt:          time.Now().UTC().Format(time.RFC3339),
	}
}

func buildDatabaseSnapshot() (*CachedInventorySnapshot, error) {
	// In a full implementation, this queries the SQLite database.
	// For now, return a nil snapshot to let the config source handle it.
	return nil, fmt.Errorf("database snapshot not yet implemented in Go sidecar")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func firstNonEmptySlice(vals ...[]string) []string {
	for _, v := range vals {
		if len(v) > 0 {
			return v
		}
	}
	return nil
}
