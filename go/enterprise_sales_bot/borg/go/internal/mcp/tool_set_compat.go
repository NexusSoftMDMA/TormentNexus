package mcp

import (
	"fmt"
	"strings"
)

// ToolSetCompatibility handles the compatibility of tool sets across MCP servers.
type ToolSetCompatibility struct{}

// NewToolSetCompatibility creates a new tool set compatibility handler.
func NewToolSetCompatibility() *ToolSetCompatibility {
	return &ToolSetCompatibility{}
}

// ToolSetCheckResult holds the result of a tool set compatibility check.
type ToolSetCheckResult struct {
	ServerName string   `json:"serverName"`
	Compatible bool     `json:"compatible"`
	Issues     []string `json:"issues,omitempty"`
	ToolCount  int      `json:"toolCount"`
}

// CheckServerCompatibility checks whether a server's tool set is compatible
// with the current environment.
func (tsc *ToolSetCompatibility) CheckServerCompatibility(server *CachedMcpServerInventory) *ToolSetCheckResult {
	result := &ToolSetCheckResult{
		ServerName: server.Name,
		ToolCount:  0,
	}

	var issues []string

	// Check for STDIO servers that require a command
	if server.Type == "STDIO" || server.Type == "" {
		if server.Command == "" {
			issues = append(issues, "STDIO server has no command configured")
		}
	}

	// Check for SSE servers that require a URL
	if server.Type == "SSE" {
		if server.URL == "" {
			issues = append(issues, "SSE server has no URL configured")
		}
	}

	// Check if server is disabled
	if !server.Enabled {
		issues = append(issues, "Server is disabled")
	}

	// Check for placeholder values
	if strings.Contains(server.Command, "YOUR_") || strings.Contains(server.URL, "YOUR_") {
		issues = append(issues, "Server has placeholder configuration values")
	}

	result.Compatible = len(issues) == 0
	result.Issues = issues
	return result
}

// ValidateToolSetName checks if a tool set name is valid.
func (tsc *ToolSetCompatibility) ValidateToolSetName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("tool set name cannot be empty")
	}
	if len(name) > 128 {
		return fmt.Errorf("tool set name too long (max 128 characters)")
	}
	for _, ch := range name {
		if ch < 32 || ch > 126 {
			return fmt.Errorf("tool set name contains invalid character")
		}
	}
	return nil
}

// MergeToolSets merges tools from multiple sets, deduplicating by namespaced name.
func (tsc *ToolSetCompatibility) MergeToolSets(sets ...[]string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, set := range sets {
		for _, tool := range set {
			if !seen[tool] {
				seen[tool] = true
				result = append(result, tool)
			}
		}
	}

	return result
}
