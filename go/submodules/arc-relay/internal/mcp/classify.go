package mcp

import "strings"

// readKeywords are tool name patterns indicating read-only operations.
var readKeywords = []string{
	"get", "list", "search", "find", "read", "fetch", "query",
	"show", "describe", "status", "check", "count", "view", "lookup",
}

// adminKeywords are tool name patterns indicating dangerous/destructive operations.
var adminKeywords = []string{
	"delete", "remove", "destroy", "drop", "purge", "reset",
	"execute", "run_command", "deploy", "install", "uninstall",
	"reboot", "shutdown", "restart", "kill",
}

// ClassifyEndpoint determines the access tier for an endpoint based on its type, name, and description.
// Resources and prompts are always "read". Tools are classified by keyword matching.
func ClassifyEndpoint(endpointType, name, description string) string {
	// Resources and prompts are inherently read-only
	if endpointType == "resource" || endpointType == "prompt" {
		return "read"
	}

	nameLower := strings.ToLower(name)

	// Check admin keywords first (more specific/dangerous)
	for _, kw := range adminKeywords {
		if containsWord(nameLower, kw) {
			return "admin"
		}
	}

	// Check read keywords
	for _, kw := range readKeywords {
		if containsWord(nameLower, kw) {
			return "read"
		}
	}

	// Description fallbacks
	descLower := strings.ToLower(description)
	if strings.Contains(descLower, "retrieves") || strings.Contains(descLower, "returns") {
		return "read"
	}
	if strings.Contains(descLower, "deletes") || strings.Contains(descLower, "destroys") || strings.Contains(descLower, "executes") {
		return "admin"
	}

	// Default to write
	return "write"
}

// containsWord checks if the name contains the keyword as a word boundary match.
// It matches: exact match, prefix (get_users), suffix (user_list), or segment (get_user_list).
func containsWord(name, keyword string) bool {
	idx := strings.Index(name, keyword)
	if idx < 0 {
		return false
	}

	// Check left boundary: start of string or underscore/hyphen
	if idx > 0 {
		c := name[idx-1]
		if c != '_' && c != '-' {
			return false
		}
	}

	// Check right boundary: end of string or underscore/hyphen
	end := idx + len(keyword)
	if end < len(name) {
		c := name[end]
		if c != '_' && c != '-' {
			return false
		}
	}

	return true
}
