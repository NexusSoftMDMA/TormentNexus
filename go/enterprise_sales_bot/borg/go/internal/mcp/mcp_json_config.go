package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// TormentNexusMcpServerEntry represents a single MCP server entry in the JSON config.
type TormentNexusMcpServerEntry struct {
	Type        string            `json:"type,omitempty"`
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	URL         string            `json:"url,omitempty"`
	Disabled    bool              `json:"disabled,omitempty"`
	Description string            `json:"description,omitempty"`
	Meta        McpServerMeta     `json:"_meta,omitempty"`
}

// McpServerMeta holds metadata for an MCP server entry.
type McpServerMeta struct {
	Description     string            `json:"description,omitempty"`
	ServerName      string            `json:"serverName,omitempty"`
	DisplayName     string            `json:"displayName,omitempty"`
	Status          string            `json:"status,omitempty"`
	Error           string            `json:"error,omitempty"`
	AlwaysOn        bool              `json:"alwaysOn,omitempty"`
	DiscoveredAt    string            `json:"discoveredAt,omitempty"`
	CacheHydratedAt string            `json:"cacheHydratedAt,omitempty"`
	ServerTags      []string          `json:"serverTags,omitempty"`
	Tools           []McpToolMetadata `json:"tools,omitempty"`
}

// McpToolMetadata holds metadata for an MCP tool within a server config.
type McpToolMetadata struct {
	Name               string      `json:"name"`
	Description        string      `json:"description,omitempty"`
	ServerDisplayName  string      `json:"serverDisplayName,omitempty"`
	ServerTags         []string    `json:"serverTags,omitempty"`
	ToolTags           []string    `json:"toolTags,omitempty"`
	SemanticGroup      string      `json:"semanticGroup,omitempty"`
	SemanticGroupLabel string      `json:"semanticGroupLabel,omitempty"`
	AdvertisedName     string      `json:"advertisedName,omitempty"`
	Keywords           []string    `json:"keywords,omitempty"`
	AlwaysOn           bool        `json:"alwaysOn,omitempty"`
	InputSchema        interface{} `json:"inputSchema,omitempty"`
}

// McpJsonConfig represents the top-level MCP JSON configuration file.
type McpJsonConfig struct {
	McpServers map[string]TormentNexusMcpServerEntry `json:"mcpServers"`
}

// LoadMcpJsonConfig loads the MCP JSON configuration from the standard paths.
// It searches for tormentnexus.config.json or mcp.json in the workspace and config directories.
func LoadMcpJsonConfig() (*McpJsonConfig, error) {
	// Search paths in order of precedence
	paths := []string{
		"tormentnexus.config.json",
		"mcp.json",
		"mcp.jsonc",
		filepath.Join(os.Getenv("HOME"), ".tormentnexus", "config.json"),
		filepath.Join(os.Getenv("USERPROFILE"), ".tormentnexus", "config.json"),
	}

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			var config McpJsonConfig
			if err := json.Unmarshal(data, &config); err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", p, err)
			}
			if config.McpServers == nil {
				config.McpServers = make(map[string]TormentNexusMcpServerEntry)
			}
			return &config, nil
		}
	}

	return &McpJsonConfig{
		McpServers: make(map[string]TormentNexusMcpServerEntry),
	}, nil
}
