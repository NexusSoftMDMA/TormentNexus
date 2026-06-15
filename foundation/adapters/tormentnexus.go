package adapters

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/NexusSoftMDMA/TormentNexus/tormentnexus"
)

type TormentNexusStatus struct {
	Assimilated       bool           `json:"assimilated"`
	TormentNexusCoreURL       string         `json:"tormentnexusCoreUrl,omitempty"`
	MemoryContext     string         `json:"memoryContext,omitempty"`
	Provider          ProviderStatus `json:"provider"`
	MCPServerNames    []string       `json:"mcpServerNames,omitempty"`
	MCPConfigPath     string         `json:"mcpConfigPath,omitempty"`
	TormentNexusRepoPath string         `json:"tormentnexusRepoPath,omitempty"`
	Warnings          []string       `json:"warnings,omitempty"`
}

type TormentNexusAdapter struct {
	tormentnexusAdapter *tormentnexus.Adapter
	workingDir  string
	homeDir     string
}

func NewTormentNexusAdapter(workingDir string) *TormentNexusAdapter {
	homeDir, _ := os.UserHomeDir()
	return &TormentNexusAdapter{
		tormentnexusAdapter: tormentnexus.NewAdapter(),
		workingDir:  workingDir,
		homeDir:     homeDir,
	}
}

func (a *TormentNexusAdapter) Status() TormentNexusStatus {
	status := TormentNexusStatus{
		Assimilated:   a.tormentnexusAdapter != nil && a.tormentnexusAdapter.Assimilated,
		MemoryContext: a.MemoryContext(),
		Provider:      BuildProviderStatus(),
	}
	if a.tormentnexusAdapter != nil {
		status.TormentNexusCoreURL = a.tormentnexusAdapter.TormentNexusCoreURL
	}
	if repoPath, ok := a.findTormentNexusRepo(); ok {
		status.TormentNexusRepoPath = repoPath
	} else {
		status.Warnings = append(status.Warnings, "adjacent tormentnexus repo not found")
	}
	if configPath, names, err := a.listMCPServers(); err == nil {
		status.MCPConfigPath = configPath
		status.MCPServerNames = names
	} else {
		status.Warnings = append(status.Warnings, err.Error())
	}
	return status
}

func (a *TormentNexusAdapter) MemoryContext() string {
	if a.tormentnexusAdapter == nil {
		return ""
	}
	return a.tormentnexusAdapter.GetMemoryContext()
}

func (a *TormentNexusAdapter) RouteMCP(request string) string {
	if a.tormentnexusAdapter == nil {
		return request
	}
	return a.tormentnexusAdapter.RouteMCP(request)
}

func (a *TormentNexusAdapter) BuildSystemContext() string {
	status := a.Status()
	parts := []string{
		"[TormentNexus Adapter]",
		fmt.Sprintf("Assimilated: %t", status.Assimilated),
	}
	if status.TormentNexusCoreURL != "" {
		parts = append(parts, fmt.Sprintf("TormentNexus Core URL: %s", status.TormentNexusCoreURL))
	}
	if status.MemoryContext != "" {
		parts = append(parts, status.MemoryContext)
	}
	if len(status.Provider.Available) > 0 {
		parts = append(parts, BuildProviderContext())
	}
	if len(status.MCPServerNames) > 0 {
		parts = append(parts, fmt.Sprintf("Configured MCP servers: %s", strings.Join(status.MCPServerNames, ", ")))
	}
	if status.TormentNexusRepoPath != "" {
		parts = append(parts, fmt.Sprintf("TormentNexus repo: %s", status.TormentNexusRepoPath))
	}
	if len(status.Warnings) > 0 {
		parts = append(parts, fmt.Sprintf("Warnings: %s", strings.Join(status.Warnings, "; ")))
	}
	return strings.Join(parts, "\n")
}

func (a *TormentNexusAdapter) listMCPServers() (string, []string, error) {
	configPath, config, err := ParseMCPConfig(a.homeDir)
	if err != nil {
		return configPath, nil, fmt.Errorf("mcp config unavailable: %w", err)
	}
	names := make([]string, 0, len(config.MCPServers))
	for name := range config.MCPServers {
		names = append(names, name)
	}
	sort.Strings(names)
	return configPath, names, nil
}

func (a *TormentNexusAdapter) findTormentNexusRepo() (string, bool) {
	candidates := []string{}
	if a.workingDir != "" {
		candidates = append(candidates,
			filepath.Join(a.workingDir, "..", "tormentnexus"),
			filepath.Join(a.workingDir, "../tormentnexus"),
		)
	}
	if a.homeDir != "" {
		candidates = append(candidates, filepath.Join(a.homeDir, "workspace", "tormentnexus"))
	}
	for _, candidate := range candidates {
		clean := filepath.Clean(candidate)
		if stat, err := os.Stat(filepath.Join(clean, "README.md")); err == nil && !stat.IsDir() {
			return clean, true
		}
	}
	return "", false
}
