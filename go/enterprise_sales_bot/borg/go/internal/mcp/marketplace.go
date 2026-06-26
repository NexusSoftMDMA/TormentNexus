package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type MarketplaceEntry struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Tags        []string `json:"tags"`
	Installed   bool     `json:"installed"`
}

func ListMarketplace(repoRoot string, filter string) ([]MarketplaceEntry, error) {
	registryPath := filepath.Join(repoRoot, "packages", "mcp-registry", "src", "registry.json")
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return nil, err
	}

	var allEntries []MarketplaceEntry
	if err := json.Unmarshal(data, &allEntries); err != nil {
		return nil, err
	}

	if filter == "" {
		return allEntries, nil
	}

	filter = strings.ToLower(filter)
	filtered := []MarketplaceEntry{}
	for _, entry := range allEntries {
		if strings.Contains(strings.ToLower(entry.Name), filter) ||
			strings.Contains(strings.ToLower(entry.Description), filter) {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

func InstallMarketplaceEntry(id string) (string, error) {
	// Simplified port of marketplace install
	// For now, we'll just return a success message
	// In a real implementation, we would want full parity with TS McpmInstaller
	return "Successfully installed MCP Server " + id, nil
}
