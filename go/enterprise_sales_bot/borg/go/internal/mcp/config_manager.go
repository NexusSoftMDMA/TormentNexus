package mcp

/**
 * @file config_manager.go
 * @module go/internal/mcp
 *
 * WHAT: Go-native implementation of the MCP configuration manager.
 * Handles reading, writing, and syncing mcp.jsonc server definitions.
 *
 * WHY: Total Autonomy — The Go sidecar must be capable of managing its own
 * MCP ecosystem and server lifecycle without the Node control plane.
 */

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type McpServerConfig struct {
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	URL         string            `json:"url,omitempty"`
	Description string            `json:"description,omitempty"`
	AlwaysOn    bool              `json:"alwaysOn,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	BearerToken string            `json:"bearerToken,omitempty"`
}

type McpConfig struct {
	McpServers map[string]McpServerConfig `json:"mcpServers"`
}

type ConfigManager struct {
	mu     sync.RWMutex
	path   string
	config McpConfig
}

func NewConfigManager(mainConfigDir string) *ConfigManager {
	path := filepath.Join(mainConfigDir, "mcp.jsonc")
	cm := &ConfigManager{
		path: path,
		config: McpConfig{
			McpServers: make(map[string]McpServerConfig),
		},
	}
	_ = cm.Load()
	return cm
}

func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Simple JSONC to JSON (strip comments)
	cleanJSON := cm.stripComments(string(data))

	var config McpConfig
	if err := json.Unmarshal([]byte(cleanJSON), &config); err != nil {
		return fmt.Errorf("failed to parse mcp.jsonc: %w", err)
	}

	if config.McpServers == nil {
		config.McpServers = make(map[string]McpServerConfig)
	}
	cm.config = config
	return nil
}

func (cm *ConfigManager) Save() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.path, data, 0644)
}

func (cm *ConfigManager) GetServers() map[string]McpServerConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	copy := make(map[string]McpServerConfig)
	for k, v := range cm.config.McpServers {
		copy[k] = v
	}
	return copy
}

func (cm *ConfigManager) AddServer(name string, cfg McpServerConfig) error {
	cm.mu.Lock()
	cm.config.McpServers[name] = cfg
	cm.mu.Unlock()
	return cm.Save()
}

func (cm *ConfigManager) RemoveServer(name string) error {
	cm.mu.Lock()
	delete(cm.config.McpServers, name)
	cm.mu.Unlock()
	return cm.Save()
}

// stripComments removes // line comments from JSONC input.
func (cm *ConfigManager) stripComments(input string) string {
	var result []byte
	lines := strings.SplitAfter(input, "\n")
	for _, line := range lines {
		// Strip // line comments (not inside strings)
		inString := false
		escaped := false
		commentStart := -1
		for i := 0; i < len(line); i++ {
			ch := line[i]
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' && inString {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = !inString
				continue
			}
			if !inString && i+1 < len(line) && ch == '/' && line[i+1] == '/' {
				commentStart = i
				break
			}
		}
		if commentStart >= 0 {
			result = append(result, []byte(line[:commentStart])...)
			result = append(result, '\n')
		} else {
			result = append(result, []byte(line)...)
		}
	}
	return string(result)
}
