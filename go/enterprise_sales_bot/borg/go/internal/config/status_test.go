package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotReportsKeyPaths(t *testing.T) {
	workspaceRoot := t.TempDir()
	configDir := filepath.Join(workspaceRoot, ".tormentnexus-go")
	mainConfigDir := filepath.Join(workspaceRoot, ".tormentnexus")
	if err := os.MkdirAll(filepath.Join(workspaceRoot, "submodules", "tormentnexus"), 0o755); err != nil {
		t.Fatalf("failed to create tormentnexus path: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, "tormentnexus.config.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to create tormentnexus config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, "mcp.jsonc"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to create mcp config: %v", err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(mainConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create main config dir: %v", err)
	}

	status := Snapshot(Config{
		Host:          "127.0.0.1",
		Port:          4300,
		ConfigDir:     configDir,
		MainConfigDir: mainConfigDir,
		WorkspaceRoot: workspaceRoot,
	})

	if !status.WorkspaceRoot.Exists {
		t.Fatalf("expected workspace root to exist")
	}
	if !status.TormentNexusSubmodule.Exists {
		t.Fatalf("expected tormentnexus submodule path to exist")
	}
	if !status.TormentNexusConfigFile.Exists || !status.MCPConfigFile.Exists {
		t.Fatalf("expected repo config files to exist, got tormentnexus=%+v mcp=%+v", status.TormentNexusConfigFile, status.MCPConfigFile)
	}
	if status.SectionedMemoryStore.Exists {
		t.Fatalf("expected sectioned memory store to be absent by default")
	}
}
