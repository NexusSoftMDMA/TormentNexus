package config

import (
	"os"
	"path/filepath"
)

type PathStatus struct {
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
}

type Status struct {
	Host                 string     `json:"host"`
	Port                 int        `json:"port"`
	BaseURL              string     `json:"baseUrl"`
	WorkspaceRoot        PathStatus `json:"workspaceRoot"`
	ConfigDir            PathStatus `json:"configDir"`
	MainConfigDir        PathStatus `json:"mainConfigDir"`
	TormentNexusConfigFile       PathStatus `json:"tormentnexusConfigFile"`
	MCPConfigFile        PathStatus `json:"mcpConfigFile"`
	GoLockPath           PathStatus `json:"goLockPath"`
	MainLockPath         PathStatus `json:"mainLockPath"`
	ImportedInstructions PathStatus `json:"importedInstructions"`
	SectionedMemoryStore PathStatus `json:"sectionedMemoryStore"`
	LegacyMemoryStore    PathStatus `json:"legacyMemoryStore"`
	TormentNexusSubmodule        PathStatus `json:"tormentnexusSubmodule"`
}

func Snapshot(cfg Config) Status {
	return Status{
		Host:                 cfg.Host,
		Port:                 cfg.Port,
		BaseURL:              cfg.BaseURL(),
		WorkspaceRoot:        buildPathStatus(cfg.WorkspaceRoot),
		ConfigDir:            buildPathStatus(cfg.ConfigDir),
		MainConfigDir:        buildPathStatus(cfg.MainConfigDir),
		TormentNexusConfigFile:       buildPathStatus(filepath.Join(cfg.WorkspaceRoot, "tormentnexus.config.json")),
		MCPConfigFile:        buildPathStatus(filepath.Join(cfg.WorkspaceRoot, "mcp.jsonc")),
		GoLockPath:           buildPathStatus(cfg.LockPath()),
		MainLockPath:         buildPathStatus(cfg.MainLockPath()),
		ImportedInstructions: buildPathStatus(cfg.ImportedInstructionsPath()),
		SectionedMemoryStore: buildPathStatus(filepath.Join(cfg.WorkspaceRoot, ".tormentnexus", "sectioned_memory.json")),
		LegacyMemoryStore:    buildPathStatus(filepath.Join(cfg.WorkspaceRoot, ".tormentnexus", "claude_mem.json")),
		TormentNexusSubmodule:        buildPathStatus(filepath.Join(cfg.WorkspaceRoot, "submodules", "tormentnexus")),
	}
}

func buildPathStatus(path string) PathStatus {
	_, err := os.Stat(path)
	return PathStatus{
		Path:   path,
		Exists: err == nil,
	}
}
