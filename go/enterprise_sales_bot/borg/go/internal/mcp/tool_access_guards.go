package mcp

import (
	"fmt"
	"strings"
	"sync"
)

// ToolAccessGuards controls access to MCP tools based on policies and permissions.
type ToolAccessGuards struct {
	mu             sync.RWMutex
	blocked        map[string]bool // namespaced tool names that are blocked
	allowed        map[string]bool // namespaced tool names explicitly allowed
	blockedServers map[string]bool // server names that are blocked
}

// NewToolAccessGuards creates a new tool access guard.
func NewToolAccessGuards() *ToolAccessGuards {
	return &ToolAccessGuards{
		blocked:        make(map[string]bool),
		allowed:        make(map[string]bool),
		blockedServers: make(map[string]bool),
	}
}

// IsAllowed checks whether a tool call is permitted.
func (tag *ToolAccessGuards) IsAllowed(namespacedName string) bool {
	tag.mu.RLock()
	defer tag.mu.RUnlock()

	serverName, _, ok := ParseNamespacedName(namespacedName)

	// Explicitly allowed tools bypass blocked lists
	if tag.allowed[namespacedName] {
		return true
	}

	// Check if tool is blocked
	if tag.blocked[namespacedName] {
		return false
	}

	// Check if server is blocked
	if ok && tag.blockedServers[serverName] {
		return false
	}

	return true
}

// BlockTool blocks a specific tool by namespaced name.
func (tag *ToolAccessGuards) BlockTool(namespacedName string) {
	tag.mu.Lock()
	defer tag.mu.Unlock()
	tag.blocked[namespacedName] = true
}

// AllowTool explicitly allows a tool, overriding any block.
func (tag *ToolAccessGuards) AllowTool(namespacedName string) {
	tag.mu.Lock()
	defer tag.mu.Unlock()
	tag.allowed[namespacedName] = true
}

// BlockServer blocks all tools from a server.
func (tag *ToolAccessGuards) BlockServer(serverName string) {
	tag.mu.Lock()
	defer tag.mu.Unlock()
	tag.blockedServers[serverName] = true
}

// UnblockTool removes a tool from the blocked list.
func (tag *ToolAccessGuards) UnblockTool(namespacedName string) {
	tag.mu.Lock()
	defer tag.mu.Unlock()
	delete(tag.blocked, namespacedName)
}

// UnblockServer removes a server from the blocked list.
func (tag *ToolAccessGuards) UnblockServer(serverName string) {
	tag.mu.Lock()
	defer tag.mu.Unlock()
	delete(tag.blockedServers, serverName)
}

// FilterTools filters a list of tool names, returning only allowed ones.
func (tag *ToolAccessGuards) FilterTools(tools []string) []string {
	var result []string
	for _, t := range tools {
		if tag.IsAllowed(t) {
			result = append(result, t)
		}
	}
	return result
}

// ValidateCall checks whether a specific tool call is permitted.
// Returns nil if allowed, or an error describing why it was rejected.
func (tag *ToolAccessGuards) ValidateCall(namespacedName string) error {
	if !tag.IsAllowed(namespacedName) {
		serverName, toolName, _ := ParseNamespacedName(namespacedName)
		if serverName != "" {
			return fmt.Errorf("tool '%s' from server '%s' is not allowed", toolName, serverName)
		}
		return fmt.Errorf("tool '%s' is not allowed", namespacedName)
	}
	return nil
}

// IsServerBlocked checks whether a server is blocked.
func (tag *ToolAccessGuards) IsServerBlocked(serverName string) bool {
	tag.mu.RLock()
	defer tag.mu.RUnlock()
	return tag.blockedServers[serverName]
}

// ListBlockedTools returns all blocked tool names.
func (tag *ToolAccessGuards) ListBlockedTools() []string {
	tag.mu.RLock()
	defer tag.mu.RUnlock()
	names := make([]string, 0, len(tag.blocked))
	for n := range tag.blocked {
		names = append(names, n)
	}
	return names
}

// IsBlockedByPrefix checks if a tool name matches any blocked prefix pattern.
func (tag *ToolAccessGuards) IsBlockedByPrefix(namespacedName string) bool {
	tag.mu.RLock()
	defer tag.mu.RUnlock()

	for blocked := range tag.blocked {
		if strings.HasPrefix(namespacedName, blocked) {
			return true
		}
	}
	return false
}
