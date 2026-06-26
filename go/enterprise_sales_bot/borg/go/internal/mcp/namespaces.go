package mcp

import "strings"

// NamespaceToolName creates a namespaced tool name from server and tool name.
// Format: "serverName__toolName" to avoid collisions across MCP servers.
func NamespaceToolName(serverName, toolName string) string {
	return serverName + "__" + toolName
}

// ParseNamespacedName splits a namespaced tool name back into server and tool.
// Returns ("serverName", "toolName", true) on success.
func ParseNamespacedName(namespaced string) (string, string, bool) {
	idx := strings.Index(namespaced, "__")
	if idx < 0 {
		return "", "", false
	}
	return namespaced[:idx], namespaced[idx+2:], true
}

// IsNamespaced checks whether a tool name contains the namespace separator.
func IsNamespaced(name string) bool {
	return strings.Contains(name, "__")
}

// SafeServerName sanitizes a server name for use in namespacing.
func SafeServerName(name string) string {
	safe := strings.NewReplacer(
		".", "_",
		"-", "_",
		" ", "_",
		"/", "_",
	).Replace(name)
	return strings.ToLower(safe)
}

// NamespaceSet represents a set of namespaced tool names for quick lookup.
type NamespaceSet struct {
	entries map[string]bool
}

// NewNamespaceSet creates a new namespace set.
func NewNamespaceSet() *NamespaceSet {
	return &NamespaceSet{entries: make(map[string]bool)}
}

// Add adds a namespaced name to the set.
func (ns *NamespaceSet) Add(name string) {
	ns.entries[name] = true
}

// Contains checks if a namespaced name is in the set.
func (ns *NamespaceSet) Contains(name string) bool {
	return ns.entries[name]
}

// Remove removes a namespaced name from the set.
func (ns *NamespaceSet) Remove(name string) {
	delete(ns.entries, name)
}

// List returns all namespaced names in the set.
func (ns *NamespaceSet) List() []string {
	var result []string
	for k := range ns.entries {
		result = append(result, k)
	}
	return result
}

// Len returns the number of entries in the set.
func (ns *NamespaceSet) Len() int {
	return len(ns.entries)
}
