package mcp

import (
	"sync"
	"time"
)

// ServerMetadataCacheEntry holds cached metadata for a single MCP server.
type ServerMetadataCacheEntry struct {
	ServerName   string    `json:"serverName"`
	DisplayName  string    `json:"displayName"`
	Tags         []string  `json:"tags"`
	ToolCount    int       `json:"toolCount"`
	LastSeen     time.Time `json:"lastSeen"`
	LastError    string    `json:"lastError,omitempty"`
	Capabilities []string  `json:"capabilities,omitempty"`
}

// ServerMetadataCache caches metadata about known MCP servers.
type ServerMetadataCache struct {
	mu      sync.RWMutex
	entries map[string]*ServerMetadataCacheEntry
	ttl     time.Duration
}

// NewServerMetadataCache creates a new server metadata cache.
func NewServerMetadataCache(ttl time.Duration) *ServerMetadataCache {
	return &ServerMetadataCache{
		entries: make(map[string]*ServerMetadataCacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a cached entry for a server, returning nil if not found or expired.
func (smc *ServerMetadataCache) Get(serverName string) *ServerMetadataCacheEntry {
	smc.mu.RLock()
	defer smc.mu.RUnlock()

	entry, ok := smc.entries[serverName]
	if !ok {
		return nil
	}

	if smc.ttl > 0 && time.Since(entry.LastSeen) > smc.ttl {
		return nil
	}

	return entry
}

// Set stores or updates a cache entry for a server.
func (smc *ServerMetadataCache) Set(serverName string, entry *ServerMetadataCacheEntry) {
	smc.mu.Lock()
	defer smc.mu.Unlock()

	entry.LastSeen = time.Now()
	smc.entries[serverName] = entry
}

// UpdateToolCount updates the tool count for a server without modifying other fields.
func (smc *ServerMetadataCache) UpdateToolCount(serverName string, count int) {
	smc.mu.Lock()
	defer smc.mu.Unlock()

	if entry, ok := smc.entries[serverName]; ok {
		entry.ToolCount = count
		entry.LastSeen = time.Now()
	}
}

// RecordError records an error for a server.
func (smc *ServerMetadataCache) RecordError(serverName string, err error) {
	smc.mu.Lock()
	defer smc.mu.Unlock()

	if entry, ok := smc.entries[serverName]; ok {
		entry.LastError = err.Error()
		entry.LastSeen = time.Now()
	} else {
		smc.entries[serverName] = &ServerMetadataCacheEntry{
			ServerName: serverName,
			LastError:  err.Error(),
			LastSeen:   time.Now(),
		}
	}
}

// Remove deletes a cached entry for a server.
func (smc *ServerMetadataCache) Remove(serverName string) {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	delete(smc.entries, serverName)
}

// Clear removes all cached entries.
func (smc *ServerMetadataCache) Clear() {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	smc.entries = make(map[string]*ServerMetadataCacheEntry)
}

// List returns all non-expired cached entries.
func (smc *ServerMetadataCache) List() []*ServerMetadataCacheEntry {
	smc.mu.RLock()
	defer smc.mu.RUnlock()

	var result []*ServerMetadataCacheEntry
	now := time.Now()
	for _, entry := range smc.entries {
		if smc.ttl > 0 && now.Sub(entry.LastSeen) > smc.ttl {
			continue
		}
		result = append(result, entry)
	}
	return result
}

// Len returns the number of cached entries (including expired).
func (smc *ServerMetadataCache) Len() int {
	smc.mu.RLock()
	defer smc.mu.RUnlock()
	return len(smc.entries)
}
