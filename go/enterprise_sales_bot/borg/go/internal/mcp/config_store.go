package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// McpConfigStore persists MCP server configurations to disk.
type McpConfigStore struct {
	mu       sync.RWMutex
	filePath string
	config   *McpJsonConfig
}

// NewMcpConfigStore creates a new config store backed by the given file path.
func NewMcpConfigStore(filePath string) *McpConfigStore {
	if filePath == "" {
		home, _ := os.UserHomeDir()
		filePath = filepath.Join(home, ".tormentnexus", "mcp-servers.json")
	}
	return &McpConfigStore{
		filePath: filePath,
	}
}

// Load loads the config from disk, returning the default if file doesn't exist.
func (cs *McpConfigStore) Load() (*McpJsonConfig, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.config != nil {
		return cs.config, nil
	}

	data, err := os.ReadFile(cs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			cs.config = &McpJsonConfig{
				McpServers: make(map[string]TormentNexusMcpServerEntry),
			}
			return cs.config, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config McpJsonConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if config.McpServers == nil {
		config.McpServers = make(map[string]TormentNexusMcpServerEntry)
	}

	cs.config = &config
	return cs.config, nil
}

// Save writes the current config to disk.
func (cs *McpConfigStore) Save() error {
	cs.mu.RLock()
	config := cs.config
	cs.mu.RUnlock()

	if config == nil {
		return fmt.Errorf("no config to save")
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(cs.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(cs.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// UpsertServer adds or updates an MCP server entry in the config.
func (cs *McpConfigStore) UpsertServer(name string, entry TormentNexusMcpServerEntry) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.config == nil {
		cs.config = &McpJsonConfig{
			McpServers: make(map[string]TormentNexusMcpServerEntry),
		}
	}

	cs.config.McpServers[name] = entry
	return cs.saveLocked()
}

// RemoveServer deletes an MCP server entry from the config.
func (cs *McpConfigStore) RemoveServer(name string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.config != nil {
		delete(cs.config.McpServers, name)
	}
	return cs.saveLocked()
}

// ListServers returns all server entries.
func (cs *McpConfigStore) ListServers() map[string]TormentNexusMcpServerEntry {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if cs.config == nil {
		return make(map[string]TormentNexusMcpServerEntry)
	}

	result := make(map[string]TormentNexusMcpServerEntry, len(cs.config.McpServers))
	for k, v := range cs.config.McpServers {
		result[k] = v
	}
	return result
}

func (cs *McpConfigStore) saveLocked() error {
	data, err := json.MarshalIndent(cs.config, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(cs.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(cs.filePath, data, 0644)
}
