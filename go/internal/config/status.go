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
	BorgConfigFile       PathStatus `json:"borgConfigFile"`
	MCPConfigFile        PathStatus `json:"mcpConfigFile"`
	GoLockPath           PathStatus `json:"goLockPath"`
	MainLockPath         PathStatus `json:"mainLockPath"`
	ImportedInstructions PathStatus `json:"importedInstructions"`
	SectionedMemoryStore PathStatus `json:"sectionedMemoryStore"`
	LegacyMemoryStore    PathStatus `json:"legacyMemoryStore"`
	BorgSubmodule        PathStatus `json:"borgSubmodule"`
}

func Snapshot(cfg Config) Status {
	return Status{
		Host:                 cfg.Host,
		Port:                 cfg.Port,
		BaseURL:              cfg.BaseURL(),
		WorkspaceRoot:        buildPathStatus(cfg.WorkspaceRoot),
		ConfigDir:            buildPathStatus(cfg.ConfigDir),
		MainConfigDir:        buildPathStatus(cfg.MainConfigDir),
		BorgConfigFile:       buildPathStatus(filepath.Join(cfg.WorkspaceRoot, "borg.config.json")),
		MCPConfigFile:        buildPathStatus(filepath.Join(cfg.WorkspaceRoot, "mcp.jsonc")),
		GoLockPath:           buildPathStatus(cfg.LockPath()),
		MainLockPath:         buildPathStatus(cfg.MainLockPath()),
		ImportedInstructions: buildPathStatus(cfg.ImportedInstructionsPath()),
		SectionedMemoryStore: buildPathStatus(filepath.Join(cfg.WorkspaceRoot, ".borg", "sectioned_memory.json")),
		LegacyMemoryStore:    buildPathStatus(filepath.Join(cfg.WorkspaceRoot, ".borg", "claude_mem.json")),
		BorgSubmodule:        buildPathStatus(filepath.Join(cfg.WorkspaceRoot, "submodules", "borg")),
	}
}

func buildPathStatus(path string) PathStatus {
	_, err := os.Stat(path)
	return PathStatus{
		Path:   path,
		Exists: err == nil,
	}
}
